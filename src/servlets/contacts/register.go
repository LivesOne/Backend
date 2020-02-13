package contacts

import (
	"server"
	"servlets/constants"
)

func RegisterHandlers() {
	server.RegisterHandler(constants.ACCOUNT_CONTACTS_LIST, new(contactListHandler))
	server.RegisterHandler(constants.ACCOUNT_CONTACTS_CREATE, new(contactCreateHandler))
	server.RegisterHandler(constants.ACCOUNT_CONTACTS_MODIFY, new(contactModifyHandler))
	server.RegisterHandler(constants.ACCOUNT_CONTACTS_DELETE, new(contactDeleteHandler))
	server.RegisterHandler(constants.ACCOUNT_CONTACTS_FIND, new(contactFindHandler))
	server.RegisterHandler(constants.ACCOUNT_CONTACTS_ADD, new(contactAddHandler))
}
