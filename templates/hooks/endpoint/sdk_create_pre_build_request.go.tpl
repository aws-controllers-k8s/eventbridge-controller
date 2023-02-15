if err = validateEndpointSpec(nil,desired.ko.Spec); err != nil {
		return nil, ackerr.NewTerminalError(err)
}
