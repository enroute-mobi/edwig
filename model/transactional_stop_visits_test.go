package model

import "testing"

func Test_TransactionalStopVisits_Find_NotFound(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	_, ok := stopVisits.Find("6ba7b814-9dad-11d1-0-00c04fd430c8")
	if ok {
		t.Errorf("Find should return false when StopVisit isn't found")
	}
}

func Test_TransactionalStopVisits_Find_Model(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	existingStopVisit := model.StopVisits().New()
	model.StopVisits().Save(&existingStopVisit)

	stopVisitId := existingStopVisit.Id()

	stopVisit, ok := stopVisits.Find(stopVisitId)
	if !ok {
		t.Errorf("Find should return true when StopVisit is found")
	}
	if stopVisit.Id() != stopVisitId {
		t.Errorf("Find should return a StopVisit with the given Id")
	}
}

func Test_TransactionalStopVisits_Find_Saved(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	existingStopVisit := stopVisits.New()
	stopVisits.Save(&existingStopVisit)

	stopVisitId := existingStopVisit.Id()

	stopVisit, ok := stopVisits.Find(stopVisitId)
	if !ok {
		t.Errorf("Find should return true when StopVisit is found")
	}
	if stopVisit.Id() != stopVisitId {
		t.Errorf("Find should return a StopVisit with the given Id")
	}
}

func Test_TransactionalStopVisits_FindAll(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	for i := 0; i < 5; i++ {
		existingStopVisit := stopVisits.New()
		stopVisits.Save(&existingStopVisit)
	}

	foundStopVisits := stopVisits.FindAll()

	if len(foundStopVisits) != 5 {
		t.Errorf("FindAll should return all stopVisits")
	}
}

func Test_TransactionalStopVisits_Save(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	stopVisit := stopVisits.New()
	objectid := NewObjectID("kind", "value")
	stopVisit.SetObjectID(objectid)
	stopVisit.VehicleJourneyId = "6ba7b814-9dad-11d1-0-00c04fd430c8"

	if success := stopVisits.Save(&stopVisit); !success {
		t.Errorf("Save should return true")
	}
	if stopVisit.Id() == "" {
		t.Errorf("New StopVisit identifier shouldn't be an empty string")
	}
	if _, ok := model.StopVisits().Find(stopVisit.Id()); ok {
		t.Errorf("StopVisit shouldn't be saved before commit")
	}
	if _, ok := model.StopVisits().FindByObjectId(objectid); ok {
		t.Errorf("StopVisit shouldn't be found by objectid before commit")
	}
	foundStopVisits := model.StopVisits().FindByVehicleJourneyId("6ba7b814-9dad-11d1-0-00c04fd430c8")
	if len(foundStopVisits) != 0 {
		t.Errorf("StopVisit shouldn't be found by VehicleJourneyId before commit")
	}
}

func Test_TransactionalStopVisits_Delete(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	existingStopVisit := model.StopVisits().New()
	objectid := NewObjectID("kind", "value")
	existingStopVisit.SetObjectID(objectid)
	existingStopVisit.VehicleJourneyId = "6ba7b814-9dad-11d1-0-00c04fd430c8"
	model.StopVisits().Save(&existingStopVisit)

	stopVisits.Delete(&existingStopVisit)

	_, ok := stopVisits.Find(existingStopVisit.Id())
	if !ok {
		t.Errorf("StopVisit should not be deleted before commit")
	}
	_, ok = stopVisits.FindByObjectId(objectid)
	if !ok {
		t.Errorf("StopVisit should be found by objectid before commit")
	}
	foundStopVisits := model.StopVisits().FindByVehicleJourneyId("6ba7b814-9dad-11d1-0-00c04fd430c8")
	if len(foundStopVisits) == 0 || foundStopVisits[0].Id() != existingStopVisit.Id() {
		t.Errorf("StopVisit should be found by VehicleJourneyId before commit")
	}
}

func Test_TransactionalStopVisits_Commit(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	// Test Save
	stopVisit := stopVisits.New()
	objectid := NewObjectID("kind", "value")
	stopVisit.SetObjectID(objectid)
	stopVisit.VehicleJourneyId = "6ba7b814-9dad-11d1-0-00c04fd430c8"
	stopVisits.Save(&stopVisit)

	// Test Delete
	existingStopVisit := model.StopVisits().New()
	secondObjectid := NewObjectID("kind", "value2")
	existingStopVisit.SetObjectID(secondObjectid)
	existingStopVisit.VehicleJourneyId = "6ba7b814-9dad-11d1-1-00c04fd430c8"
	model.StopVisits().Save(&existingStopVisit)
	stopVisits.Delete(&existingStopVisit)

	stopVisits.Commit()

	if _, ok := model.StopVisits().Find(stopVisit.Id()); !ok {
		t.Errorf("StopVisit should be saved after commit")
	}
	if _, ok := model.StopVisits().FindByObjectId(objectid); !ok {
		t.Errorf("StopVisit should be found by objectid after commit")
	}
	foundStopVisits := model.StopVisits().FindByVehicleJourneyId("6ba7b814-9dad-11d1-0-00c04fd430c8")
	if len(foundStopVisits) == 0 || foundStopVisits[0].Id() != stopVisit.Id() {
		t.Errorf("StopVisit should be found by VehicleJourneyId after commit")
	}

	if _, ok := model.StopVisits().Find(existingStopVisit.Id()); ok {
		t.Errorf("StopVisit should be deleted after commit")
	}
	if _, ok := model.StopVisits().FindByObjectId(secondObjectid); ok {
		t.Errorf("StopVisit shouldn't be found by objectid after commit")
	}
	foundStopVisits = model.StopVisits().FindByVehicleJourneyId("6ba7b814-9dad-11d1-1-00c04fd430c8")
	if len(foundStopVisits) != 0 {
		t.Errorf("StopVisit shouldn't be found by VehicleJourneyId after commit")
	}
}

func Test_TransactionalStopVisits_Rollback(t *testing.T) {
	model := NewMemoryModel()
	stopVisits := NewTransactionalStopVisits(model)

	stopVisit := stopVisits.New()
	objectid := NewObjectID("kind", "value")
	stopVisit.SetObjectID(objectid)
	stopVisit.VehicleJourneyId = "6ba7b814-9dad-11d1-0-00c04fd430c8"
	stopVisits.Save(&stopVisit)

	stopVisits.Rollback()
	stopVisits.Commit()

	if _, ok := model.StopVisits().Find(stopVisit.Id()); ok {
		t.Errorf("StopVisit should not be saved with a rollback")
	}
	if _, ok := model.StopVisits().FindByObjectId(objectid); ok {
		t.Errorf("StopVisit shouldn't be found by objectid with a rollback")
	}
	foundStopVisits := model.StopVisits().FindByVehicleJourneyId("6ba7b814-9dad-11d1-0-00c04fd430c8")
	if len(foundStopVisits) != 0 {
		t.Errorf("StopVisit shouldn't be found by VehicleJourneyId with a rollback")
	}
}
