package model

//create table goods
//(
//id          int auto_increment comment '自增ID'
//primary key,
//good_id     varchar(64)  default ''                not null comment '商品ID',
//user_id     int                                    not null comment '用户ID（包括用户ID、教练ID、俱乐部ID、官方ID）',
//user_type   int          default 0                 not null comment '用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）',
//course_id   varchar(64)                            null comment '单次课的课程ID',
//title       varchar(255) default ''                not null comment '商品标题',
//cover_url   varchar(255)                           null comment '封面',
//detail      text                                   null comment '商品详情',
//pack        int          default 0                 not null comment '是否为打包课程（0：否，1：打包）',
//on_shelf    int          default 0                 not null comment '是否上架，0-下架，1-上架',
//teach_time  int          default 0                 not null comment '教学时长',
//teach_money int          default 0                 not null comment '教学费用',
//club_money int          default 0                 null comment '推荐费（俱乐部的课程才有）',
//area_money  int          default 0                 not null comment '场地费用',
//fault_money int          default 0                 not null comment '有责取消费',
//discount    int          default 10                not null comment '折扣（打包课才设置）',
//stack       int          default -1                not null comment '库存（-1：不限库存）',
//state       int          default 0                 not null comment '状态，0-正常， 1-删除',
//created_at  timestamp    default CURRENT_TIMESTAMP not null comment '创建时间',
//updated_at  timestamp    default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
//constraint good_id_pk
//unique (good_id),
//constraint user_id_pk
//unique (user_id)
//)
//comment '商品表' collate = utf8mb4_general_ci;

type Goods struct {
	ID            int                    `json:"id" gorm:"column:id"`                          // 自增ID
	GoodID        string                 `json:"good_id" gorm:"column:good_id"`                // 商品ID
	UserID        string                 `json:"user_id" gorm:"column:user_id"`                // 用户ID（包括用户ID、教练ID、俱乐部ID、官方ID）
	UserType      int                    `json:"user_type" gorm:"column:user_type"`            // 用户类型（1：普通用户、2：教练、3：俱乐部、4：官方）
	CourseID      string                 `json:"course_id" gorm:"column:course_id"`            // 单次课的课程ID
	Title         string                 `json:"title" gorm:"column:title"`                    // 商品标题
	CoverUrl      string                 `json:"cover_url" gorm:"column:cover_url"`            // 封面
	Detail        string                 `json:"detail" gorm:"column:detail"`                  // 商品详情
	Pack          int                    `json:"pack" gorm:"column:pack"`                      // 是否为打包课程（0：否，1：打包）
	OnShelf       int                    `json:"on_shelf" gorm:"column:on_shelf; default:0"`   // 是否上架，0-下架，1-上架
	PointsDeduct  int                    `json:"points_deduct" gorm:"column:points_deduct"`    // 积分抵扣（0：否，1：抵扣）
	TeachTime     int                    `json:"teach_time" gorm:"column:teach_time"`          // 教学时长
	TeachMoney    int64                  `json:"teach_money" gorm:"column:teach_money"`        // 教学费用
	ClubMoney     int64                  `json:"club_money" gorm:"column:club_money"`          // 推荐费（俱乐部的课程才有）
	AreaMoney     int64                  `json:"area_money" gorm:"column:area_money"`          // 场地费用
	FaultMoney    int64                  `json:"fault_money" gorm:"column:fault_money"`        // 有责取消费
	TotalMoney    int64                  `json:"total_money" gorm:"column:total_money"`        // 总费用
	Discount      int                    `json:"discount" gorm:"column:discount; default:100"` // 折扣（打包课才设置）
	DiscountMoney int64                  `json:"discount_money" gorm:"column:discount_money"`  // 折扣金额
	FinishedCnt   int                    `gorm:"column:finished_cnt" json:"finished_cnt"`
	UnFinishedCnt int                    `gorm:"column:unfinished_cnt" json:"unfinished_cnt"`
	CanceledCnt   int                    `gorm:"column:canceled_cnt" json:"canceled_cnt"`
	Stack         int                    `json:"stack" gorm:"column:stack;default:-1"` // 库存（-1：不限库存）
	State         int                    `json:"-" gorm:"column:state"`                // 状态，0-正常， 1-删除
	CreatedAt     LocalTime              `json:"created_at" gorm:"column:created_at"`  // 创建时间
	UpdatedAt     LocalTime              `json:"updated_at" gorm:"column:updated_at"`  // 更新时间
	CourseTags    []*CoursesTags         `gorm:"foreignKey:CourseID;references:CourseID" json:"-"`
	GoodsCourses  []*GoodsCourses        `gorm:"foreignKey:GoodID;references:GoodID" json:"goods_courses,omitempty"`
	Tags          []*Tags                `gorm:"-" json:"tags,omitempty"`
	PayData       QueryGoodBeforeBuyResp `gorm:"-" json:"pay_data"`
}

type QueryGoodBeforeBuyResp struct {
	PayFee    int64 `json:"pay_fee"`
	PointsFee int64 `json:"points_fee"`
}

func (m *Goods) TableName() string {
	return "goods"
}

const PackNo = 0  // 非打包课程
const PackYes = 1 // 打包课程

const OnShelfNo = 0      // 下架
const OnShelfYes = 1     // 上架
const OnShelfAdminNo = 2 // 管理台下架，无法上架
const PointsDeductNo = 0
const PointsDeductYes = 1
