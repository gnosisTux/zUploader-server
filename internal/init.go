package internal

func Init() {
	InitDB(ConfigData.AuditDB)
	schedulePurge()
}
