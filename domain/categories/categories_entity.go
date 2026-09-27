package categories

import "time"

// Type ประเภทของหมวดหมู่
type Type string

const (
	TypeIncome  Type = "income"
	TypeExpense Type = "expense"
	TypeBoth    Type = "both"
)

// Valid ตรวจสอบว่าเป็นประเภทที่ระบบรองรับ
func (t Type) Valid() bool {
	switch t {
	case TypeIncome, TypeExpense, TypeBoth:
		return true
	}
	return false
}

// Category entity ของหมวดหมู่ที่ผู้ใช้เป็นเจ้าของ
type Category struct {
	ID            int
	UserProfileID int
	Name          string
	Color         string
	Icon          string
	Type          Type
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// MasterCategory entity ของหมวดหมู่ตั้งต้นของระบบ
type MasterCategory struct {
	ID        int
	Name      string
	Color     string
	Icon      string
	Type      Type
	CreatedAt time.Time
	UpdatedAt time.Time
}
