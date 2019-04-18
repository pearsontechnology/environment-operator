package diff

var changeMap map[string]string

func newChangeMap() {
	changeMap = make(map[string]string)
}

func addServiceChange(svc, diff string) {
	changeMap[svc] = diff
}

func ServiceChanged(serviceName string) bool {
	_, serviceChangeExists := changeMap[serviceName]

	if serviceChangeExists {
		return true
	}
	return false
}

func Changes() map[string]string {
	return changeMap
}

