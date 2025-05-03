package sneakers

import (
	"context"
	"strconv"
	"strings"
	"time"
)

type service struct {
	Repository
	timeout time.Duration
}

func NewService(repository Repository) Service {
	return &service{
		repository,
		time.Duration(2) * time.Second,
	}
}

func (s *service) AddSneaker(c context.Context, req *CreateSneakerReq) (*Sneaker, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	sn := &Sneaker{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		ImageUrl:    req.ImageUrl,
	}

	r, err := s.Repository.AddSneaker(ctx, sn)
	if err != nil {
		return nil, err
	}
	res := &Sneaker{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		Price:       r.Price,
		ImageUrl:    r.ImageUrl,
	}

	return res, nil
}

func (s *service) DeleteSneaker(c context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.Repository.DeleteSneaker(ctx, id)
}

func (s *service) GetAllSneakers(c context.Context) ([]*Sneaker, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.Repository.GetAllSneakers(ctx)
}

func (s *service) GetSneakerByID(c context.Context, id int64) (*Sneaker, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.Repository.GetSneakerByID(ctx, id)
}

func (s *service) GetSneakersByIDs(c context.Context, ids []int64) ([]*Sneaker, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	return s.Repository.GetSneakersByIDs(ctx, ids)
}

// ParseIDsString парсит строку с ID товаров, разделенных запятыми
func (s *service) ParseIDsString(idsString string) ([]int64, error) {
	// Если строка пустая, возвращаем пустой массив
	if idsString == "" {
		return []int64{}, nil
	}

	// Разделяем строку по запятым
	idStrings := strings.Split(idsString, ",")
	
	// Создаем массив для результата
	ids := make([]int64, 0, len(idStrings))
	
	// Преобразуем каждую строку в число
	for _, idStr := range idStrings {
		// Пропускаем пустые строки
		if idStr == "" {
			continue
		}
		
		// Преобразуем строку в число
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			return nil, err
		}
		
		// Добавляем число в результат
		ids = append(ids, id)
	}
	
	return ids, nil
}