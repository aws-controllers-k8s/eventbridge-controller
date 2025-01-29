if archiveInTerminalState(latest) {
	msg := fmt.Sprintf("Archive is in status %q", *latest.ko.Status.State)
	ackcondition.SetTerminal(desired, corev1.ConditionTrue, &msg, nil)
	ackcondition.SetSynced(desired, corev1.ConditionTrue, nil, nil)
	return desired, nil
}

if archiveModifying(latest) {
	msg := fmt.Sprintf("Archive is in status %q", *latest.ko.Status.State)
	ackcondition.SetSynced(desired, corev1.ConditionFalse, &msg, nil)
	return desired, requeueWaitUntilCanModify(latest)
}

if !archiveAvailable(latest) {
	msg := fmt.Sprintf("Archive is not available for modification in status %q", *latest.ko.Status.State)
	ackcondition.SetSynced(desired, corev1.ConditionFalse, &msg, nil)
	return desired, requeueWaitUntilCanModify(latest)
}
