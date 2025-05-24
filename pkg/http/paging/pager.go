package paging

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math"
	"reflect"
	"strings"
)

const (
	defaultPageSize = 100
	maxPageSize     = 500
)

// Pager represents a object that support paginate data in DB
// also parse request from client via gin.Context
type Pager struct {
	Page           int         `json:"page" form:"page"`
	PageSize       int         `json:"page_size" form:"page_size"`
	Sort           string      `json:"sort" form:"sort"`
	TotalRows      int64       `json:"total"`
	SortableFields []string    `json:"sortable_fields"`
	Metadata       interface{} `json:"metadata"`
}

type SortableFieldsGetter interface {
	GetSortableFields() []string
}

// NewPagerWithGinCtx initializes a new Pager from gin context by reading in order query, body request
func NewPagerWithGinCtx(c *gin.Context) *Pager {
	pg := &Pager{}
	if err := c.ShouldBind(pg); err != nil {
		return nil
	}
	return pg
}

func (p *Pager) GetPage() int {
	if p.Page == 0 {
		return 1
	}
	return p.Page
}

func (p *Pager) GetOffset() int {
	return (p.GetPage() - 1) * p.PageSize
}

func (p *Pager) GetPageSize() int {
	if p.PageSize == 0 {
		return defaultPageSize
	}
	if p.PageSize > maxPageSize {
		return maxPageSize
	}
	return p.PageSize
}

// zerost is a empty struct (zero memory allowed), used for indexing map
type zerost struct{}

// GetOrder parses the order field then return Gorm order format
//
//	eg. "name,-age" => "name asc, age desc"
func (p *Pager) GetOrder(sortableFields []string) string {
	rawSegments := strings.Split(p.Sort, ",")
	var finalSortFields []string

	// sortable fields index
	var sortableFieldsIdx = map[string]zerost{}
	for _, field := range sortableFields {
		sortableFieldsIdx[field] = zerost{}
	}

	for _, segment := range rawSegments {
		segment = strings.TrimSpace(segment)

		var (
			fieldName string
			direction = "asc"
		)

		// convert :
		// 	-field -> field desc
		//	field -> field asc
		if strings.HasPrefix(segment, "-") {
			fieldName = segment[1:]
			direction = "desc"
		} else {
			fieldName = segment
		}

		if _, ok := sortableFieldsIdx[fieldName]; ok {
			finalSortFields = append(finalSortFields, fieldName+" "+direction)
		}
	}

	return strings.Join(finalSortFields, ", ")
}

func (p *Pager) GetTotalPages() int {
	return int(math.Ceil(float64(p.TotalRows) / float64(p.GetPageSize())))
}

func (p *Pager) CanNext() bool {
	canNext := (p.Page * p.PageSize) < int(p.TotalRows)
	return canNext
}

func (p *Pager) CanPre() bool {
	return p.Page > 1
}

func (p *Pager) TraceID() string {
	TradeId, _ := uuid.NewUUID()
	return TradeId.String()
}

// DoQuery The execution will stop on count error then return that transaction
func (p *Pager) DoQuery(value interface{}, db *gorm.DB) *gorm.DB {
	var (
		totalRows int64
		tx        *gorm.DB
	)
	if tx = db.Count(&totalRows); tx.Error != nil {
		return tx
	}
	p.TotalRows = totalRows

	sortableFields := p.SortableFields
	if len(p.SortableFields) == 0 {
		sortableFields = p.resolveSortableFields(value)
	}
	order := p.GetOrder(sortableFields)

	tx = db.Offset(p.GetOffset()).Limit(p.GetPageSize())
	if order != "" {
		tx = tx.Order(order)
	}

	return tx.Find(value)
}

func (p *Pager) resolveSortableFields(value interface{}) []string {
	var fields []string
	refType := reflect.TypeOf(value)
	for refType.Kind() == reflect.Ptr || refType.Kind() == reflect.Slice {
		refType = refType.Elem()
	}
	ptr := reflect.New(refType)
	if getter, ok := ptr.Interface().(SortableFieldsGetter); ok {
		fields = getter.GetSortableFields()
	}
	return fields
}

// Phần tính toán dành riêng cho GetListOwnerTruck
func (p *Pager) PageCount() int {

	rs := int(p.TotalRows) / p.PageSize
	if int(p.TotalRows)%p.PageSize != 0 {
		rs++
	}
	return rs
}

func (p *Pager) DoQueryListOwnerTruck(value interface{}, count int64, db *gorm.DB) *gorm.DB {
	var (
		tx *gorm.DB
	)
	p.TotalRows = count

	sortableFields := p.SortableFields
	if len(p.SortableFields) == 0 {
		sortableFields = p.resolveSortableFields(value)
	}
	order := p.GetOrder(sortableFields)

	tx = db.Offset(p.GetOffset()).Limit(p.GetPageSize())
	if order != "" {
		tx = tx.Order(order)
	}

	return tx.Find(value)
}

func (p *Pager) DoQueryListTruckAvailableWithDriver(value interface{}, count int64, db *gorm.DB) *gorm.DB {
	var (
		tx *gorm.DB
	)
	p.TotalRows = count

	sortableFields := p.SortableFields
	if len(p.SortableFields) == 0 {
		sortableFields = p.resolveSortableFields(value)
	}
	order := p.GetOrder(sortableFields)

	tx = db.Offset(p.GetOffset()).Limit(p.GetPageSize())
	if order != "" {
		tx = tx.Order(order)
	}

	return tx.Find(value)
}

func (p *Pager) DoQueryRawSql(value interface{}, db *gorm.DB, sql string, args ...interface{}) *gorm.DB {
	var (
		totalRows int64
		tx        *gorm.DB
	)

	if strings.HasSuffix(sql, ";") {
		sql = strings.TrimSuffix(sql, ";")
	}

	countSql := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) AS subquery`, sql)
	if tx = db.Raw(countSql, args...).Scan(&totalRows); tx.Error != nil {
		return tx
	}
	p.TotalRows = totalRows

	sortableFields := p.SortableFields
	if len(p.SortableFields) == 0 {
		sortableFields = p.resolveSortableFields(value)
	}
	order := p.GetOrder(sortableFields)

	tx = db.Offset(p.GetOffset()).Limit(p.GetPageSize())
	if order != "" {
		sql += " ORDER BY " + order
	}

	// Use Raw and Scan instead of Find
	return tx.Raw(sql, args...).Scan(value)
}
