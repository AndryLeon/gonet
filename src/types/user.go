package types

type User struct {
	Id             int32  // 用户id
	ClanId         int32  // 部落ID
	ClanMsgId      uint32 // 收到的群消息最大编号
	Name           string // 用户名
	Pass           []byte // 密码(MD5 Hash)
	Score          int32  // 分数
	ProtectTimeout int64  // 护盾截止时间
	Mac            string // 玩家MAC地址
	CountryCode    string // 国家代码
	Language       string // 界面语言
	DeviceType     string // 设备类型
	CreatedAt      int64  // 注册时间
}
