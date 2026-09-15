package repository

import "gorm.io/gorm/clause"

// clauseLocking 返回 SELECT ... FOR UPDATE 子句（并发下单防超卖）。
func clauseLocking() clause.Locking {
	return clause.Locking{Strength: "UPDATE"}
}
