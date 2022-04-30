package database

type Debug struct {
	Model
	DailyCount uint64
	WorkCount  uint64
}

func (Debug) TableName() string {
	return "debug"
}
