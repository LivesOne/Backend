package token

// const (
// 	TK_ERR_OK        = 0
// 	TK_ERR_DB        = -1
// 	TK_ERR_DUPLICATE = -2
// 	TK_ERR_NOTEXIST  = -3
// )

type Database interface {
	Open(conf map[string]string)
	Insert(hash, uid, key, token string, expire int) int
	Update(hash, key string, expire int) int
	Delete(hash string) int

	GetUID(hash string) (string, int)
	GetKey(hash string) (string, int)
	GetToken(hash string) (string, int)
	GetAll(hash string) (uid, key, token string, ret int)
}
