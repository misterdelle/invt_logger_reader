package ports

type Device interface {
	Name() string
	Query() (map[string]interface{}, error)
	QueryStation() (map[string]interface{}, error)
	QueryEnergyTodayTotals() (map[string]interface{}, error)
	QueryGridOutput() (map[string]interface{}, error)
	QueryInverterInfo() (map[string]interface{}, error)
	QueryLoadInfo() (map[string]interface{}, error)
	QueryBatteryOutput() (map[string]interface{}, error)
	QueryPVOutput() (map[string]interface{}, error)
}
