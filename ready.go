package nautobotor

// Ready implements the ready.Readiness interface, once this flips to true CoreDNS
// assumes this plugin is ready for queries; it is not checked again.
func (e Nautobotor) Ready() bool {
	log.Debug("The Nautobotor module is ready")
	return true
}
