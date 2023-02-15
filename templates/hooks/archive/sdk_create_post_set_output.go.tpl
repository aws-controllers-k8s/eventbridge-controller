if !archiveAvailable(&resource{ko}) {
	ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
} else {
	ackcondition.SetSynced(&resource{ko}, corev1.ConditionTrue, nil, nil)
}
