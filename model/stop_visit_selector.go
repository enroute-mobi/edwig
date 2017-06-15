package model

import "time"

type StopVisitSelector func(StopVisit) bool

func StopVisitSelectorByTime(startTime, endTime time.Time) StopVisitSelector {
	return func(stopVisit StopVisit) bool {
		if stopVisit.ReferenceTime().IsZero() || stopVisit.ReferenceTime().Before(startTime) || stopVisit.ReferenceTime().After(endTime) {
			return false
		}
		return true
	}
}

func StopVisitSelectorByLine(objectid ObjectID) StopVisitSelector {
	return func(stopVisit StopVisit) bool {
		lineObjectid, ok := stopVisit.VehicleJourney().Line().ObjectID(objectid.Kind())
		if ok {
			return lineObjectid.Value() == objectid.Value()
		}
		return false
	}
}

func CompositeStopVisitSelector(selectors []StopVisitSelector) StopVisitSelector {
	return func(stopVisit StopVisit) bool {
		for _, selector := range selectors {
			if !selector(stopVisit) {
				return false
			}
		}
		return true
	}
}
