package models

// для работы между сервисным слоем и слоем работы с данными мы заводим модели
// которые будут доступны любому слою
type App struct {
	ID     int
	Name   string
	Secret string //подписывать токены
}
