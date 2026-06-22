package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/paulwwyvern/gophermart/internal/model"
	"github.com/paulwwyvern/gophermart/internal/model/dto"
	"github.com/paulwwyvern/gophermart/internal/model/errs"
	"github.com/paulwwyvern/gophermart/pkg/luhn"
	"github.com/shopspring/decimal"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, userId int64, number string) (orderUserId int64, err error)
	GetOrdersByUserID(ctx context.Context, userId int64) ([]*model.Order, error)
}

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, userId int64, number string) error {

	isValid := luhn.Validate(number)
	if isValid != 0 {
		if isValid < 0 {
			return errs.ErrOrderNumberInvalid
		}
		return errs.ErrOrderNumberUnprocessable
	}

	orderUserId, err := s.repo.CreateOrder(ctx, userId, number)
	if err != nil {
		if errors.Is(err, errs.ErrOrderAlreadyExists) {
			if orderUserId != userId {
				return errs.ErrOrderConflict
			}
			return errs.ErrOrderAlreadyUploaded
		}

		return err
	}

	return nil
}

func (s *OrderService) GetOrdersByUserID(ctx context.Context, userId int64) ([]*dto.GetOrdersResponse, error) {
	orders, err := s.repo.GetOrdersByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}

	res := make([]*dto.GetOrdersResponse, 0, len(orders))

	for _, order := range orders {
		res = append(res, &dto.GetOrdersResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	return res, nil
}

type OrderAccrualRepository interface {
	GetOrderStatus(ctx context.Context, number string) (string, decimal.Decimal, error)
}

type OrderStatusRepository interface {
	BeginTx(ctx context.Context) (context.Context, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error
	AddUserBalanceByID(ctx context.Context, userID int64, add decimal.Decimal) error
	GetNewOrProcessingOrderNumbers(ctx context.Context, batchSize int) ([]*model.ProcessingOrder, error)
	SetOrderStatus(ctx context.Context, number string, status string, accrual decimal.Decimal) error
}

type OrderWorkerPool struct {
	accrualRepository OrderAccrualRepository
	statusRepository  OrderStatusRepository

	isCoolDown atomic.Bool

	numWorkers int
	batchSize  int

	retryAfterDefault int

	closeCh chan struct{}
}

func NewOrderWorkerPool(accrualRepository OrderAccrualRepository, statusRepository OrderStatusRepository, numWorkers int, batchSize int, retryAfterDefault int) *OrderWorkerPool {
	return &OrderWorkerPool{
		accrualRepository: accrualRepository,
		statusRepository:  statusRepository,

		numWorkers:        numWorkers,
		batchSize:         batchSize,
		retryAfterDefault: retryAfterDefault,
	}
}

func (p *OrderWorkerPool) Run(ctx context.Context, logger *slog.Logger, interval time.Duration) {
	p.closeCh = make(chan struct{})

	var wg sync.WaitGroup
	sigCh := make(chan struct{}, p.numWorkers)
	errCh := make(chan error, 1)

	go func() {
		p.runner(ctx, interval, sigCh)
	}()

	wg.Add(p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		go func() {
			defer wg.Done()
			p.worker(ctx, logger, sigCh, errCh)
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	go func() {
		p.errorWorker(logger, errCh)
		close(p.closeCh)
	}()
}

func (p *OrderWorkerPool) Wait() {
	<-p.closeCh
}

// runner отправляет раз в установленный интервал сигнал что нужно проверить статусы.
//
// Если ранее сработал RateLimiter, то сигналы не отправляются
func (p *OrderWorkerPool) runner(ctx context.Context, interval time.Duration, sigCh chan<- struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(sigCh)
	for {
		select {
		case <-ticker.C:
			if p.isCoolDown.Load() {
				continue
			}
			for i := 0; i < p.numWorkers; i++ {
				sigCh <- struct{}{}
			}
		case <-ctx.Done():
			return
		}
	}
}

// worker получив сигнал:
//
// 1) выгребает из бд batchSize заказов, которые в статусе NEW или PROCESSING для обновления статуса
//
// 2) Для каждого заказа идёт в сторонний сервис, где узнаёт его статус
//
// 3) Если обработка заказа закончена и пользователь должен получить баллы --- они начисляются
func (p *OrderWorkerPool) worker(ctx context.Context, logger *slog.Logger, sigCh <-chan struct{}, errCh chan<- error) {
	for _ = range sigCh {
		err := p.getOrderStatus(ctx, logger, p.batchSize)
		if err != nil {
			errCh <- err
		}
	}
}

// errorWorker получает от воркеров ошибки, если что то пошло не так b логирует их
//
// В частности, если сработал RateLimiter, то ставится специальный флаг, чтобы runner не отправлял новые команды
// и через указанный промежуток времени флаг снимается
func (p *OrderWorkerPool) errorWorker(logger *slog.Logger, errCh <-chan error) {
	for err := range errCh {
		if err == nil {
			continue
		}
		var errTooManyRequest *errs.ErrAccrualTooManyRequests
		if errors.As(err, &errTooManyRequest) {
			retryAfter := errTooManyRequest.RetryAfter
			if retryAfter == 0 {
				retryAfter = p.retryAfterDefault
			}

			if !p.isCoolDown.CompareAndSwap(false, true) {
				continue
			}

			go func() {
				time.Sleep(time.Duration(retryAfter) * time.Second)
				p.isCoolDown.Store(false)
			}()
		}

		logger.Info("Get order Worker Pool error", slog.String("error", err.Error()))
	}
}

func (p *OrderWorkerPool) getOrderStatus(ctx context.Context, logger *slog.Logger, batchSize int) error {
	ctxTx, err := p.statusRepository.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer p.statusRepository.RollbackTx(ctxTx)

	// получаем из нашей бд список заказов, получаем batchSize NEW or PROCESSING заказов,
	// которые дольше всего не обновлялись
	orders, err := p.statusRepository.GetNewOrProcessingOrderNumbers(ctxTx, batchSize)

	if err != nil {
		return err
	}
	var errTooManyRequests *errs.ErrAccrualTooManyRequests
	for _, order := range orders {
		userID := order.UserID
		number := order.OrderNumber

		// Получаем статус заказа
		status, accrual, err := p.accrualRepository.GetOrderStatus(ctxTx, number)
		if err != nil {
			// при RateLimiter прекращает обработку следующих заказов, но уже обработанные сохраняем
			if errors.As(err, &errTooManyRequests) {
				break
			}
			// Если заказ не зареган
			if errors.Is(err, errs.ErrAccrualNotRegistered) {
				continue
			}
			return err
		}

		// только зарегестрированные заказы у себя переводим в PROCESSING
		if status == "REGISTERED" {
			status = "PROCESSING"
		}

		// уведомляем про заказы, для которых завершили расчёт вознаграждения или отказали в нём
		if status == "PROCESSED" || status == "INVALID" {
			logger.Info("Complete processing order",
				slog.String("order number", number),
				slog.String("status", status),
				slog.String("accrual", accrual.String()),
			)
		}

		// Если завершили расчёт вознаграждения, то передаём юзеру его баллы
		if status == "PROCESSED" {
			err = p.statusRepository.AddUserBalanceByID(ctxTx, userID, accrual)
			if err != nil {
				return err
			}
			logger.Info("Add User Balance",
				slog.Int64("user id", userID),
				slog.String("order number", number),
				slog.String("accrual", accrual.String()),
			)

		} else {
			accrual = decimal.Zero
		}

		//Устанавливаем статус заказа + обновляем время обработки
		err = p.statusRepository.SetOrderStatus(ctxTx, number, status, accrual)
		if err != nil {
			return err
		}
	}
	err = p.statusRepository.CommitTx(ctxTx)
	if err != nil {
		return err
	}

	if errTooManyRequests != nil {
		return errTooManyRequests
	}

	return nil
}
