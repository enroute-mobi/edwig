package model

import (
	"encoding/json"
	"sync"
)

type VehicleId ModelId

type Vehicle struct {
	ObjectIDConsumer

	model Model

	id               VehicleId
	VehicleJourneyId VehicleJourneyId `json:",omitempty"`

	Longitude float64 `json:",omitempty"`
	Latitude  float64 `json:",omitempty"`
	Bearing   float64 `json:",omitempty"`
}

func NewVehicle(model Model) *Vehicle {
	vehicle := &Vehicle{
		model: model,
	}
	vehicle.objectids = make(ObjectIDs)
	return vehicle
}

func (vehicle *Vehicle) modelId() ModelId {
	return ModelId(vehicle.id)
}

func (vehicle *Vehicle) Id() VehicleId {
	return vehicle.id
}

func (vehicle *Vehicle) Save() (ok bool) {
	ok = vehicle.model.Vehicles().Save(vehicle)
	return
}

func (vehicle *Vehicle) VehicleJourney() *VehicleJourney {
	vehicleJourney, ok := vehicle.model.VehicleJourneys().Find(vehicle.VehicleJourneyId)
	if !ok {
		return nil
	}
	return &vehicleJourney
}

func (vehicle *Vehicle) MarshalJSON() ([]byte, error) {
	type Alias Vehicle
	aux := struct {
		Id        VehicleId
		ObjectIDs ObjectIDs `json:",omitempty"`
		*Alias
	}{
		Id:    vehicle.id,
		Alias: (*Alias)(vehicle),
	}

	if !vehicle.ObjectIDs().Empty() {
		aux.ObjectIDs = vehicle.ObjectIDs()
	}

	return json.Marshal(&aux)
}

func (vehicle *Vehicle) UnmarshalJSON(data []byte) error {
	type Alias Vehicle
	aux := &struct {
		ObjectIDs map[string]string
		*Alias
	}{
		Alias: (*Alias)(vehicle),
	}
	err := json.Unmarshal(data, aux)
	if err != nil {
		return err
	}

	if aux.ObjectIDs != nil {
		vehicle.ObjectIDConsumer.objectids = NewObjectIDsFromMap(aux.ObjectIDs)
	}

	return nil
}

type MemoryVehicles struct {
	UUIDConsumer

	model *MemoryModel

	mutex        *sync.RWMutex
	byIdentifier map[VehicleId]*Vehicle
	byObjectId   *ObjectIdIndex
}

type Vehicles interface {
	UUIDInterface

	New() Vehicle
	Find(id VehicleId) (Vehicle, bool)
	FindByObjectId(objectid ObjectID) (Vehicle, bool)
	FindAll() []Vehicle
	Save(vehicle *Vehicle) bool
	Delete(vehicle *Vehicle) bool
}

func NewMemoryVehicles() *MemoryVehicles {
	return &MemoryVehicles{
		mutex:        &sync.RWMutex{},
		byIdentifier: make(map[VehicleId]*Vehicle),
		byObjectId:   NewObjectIdIndex(),
	}
}

func (manager *MemoryVehicles) New() Vehicle {
	vehicle := NewVehicle(manager.model)
	return *vehicle
}

func (manager *MemoryVehicles) Find(id VehicleId) (Vehicle, bool) {
	if id == "" {
		return Vehicle{}, false
	}

	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	vehicle, ok := manager.byIdentifier[id]
	if ok {
		return *vehicle, true
	} else {
		return Vehicle{}, false
	}
}

func (manager *MemoryVehicles) FindByObjectId(objectid ObjectID) (Vehicle, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	id, ok := manager.byObjectId.Find(objectid)
	if ok {
		return *manager.byIdentifier[VehicleId(id)], true
	}
	return Vehicle{}, false
}

func (manager *MemoryVehicles) FindAll() (vehicles []Vehicle) {
	manager.mutex.RLock()

	if len(manager.byIdentifier) == 0 {
		manager.mutex.RUnlock()
		return
	}
	for _, vehicle := range manager.byIdentifier {
		vehicles = append(vehicles, *vehicle)
	}
	manager.mutex.RUnlock()

	return
}

func (manager *MemoryVehicles) Save(vehicle *Vehicle) bool {
	if vehicle.Id() == "" {
		vehicle.id = VehicleId(manager.NewUUID())
	}

	manager.mutex.Lock()

	vehicle.model = manager.model
	manager.byIdentifier[vehicle.Id()] = vehicle
	manager.byObjectId.Index(vehicle)

	manager.mutex.Unlock()

	return true
}

func (manager *MemoryVehicles) Delete(vehicle *Vehicle) bool {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	delete(manager.byIdentifier, vehicle.Id())
	manager.byObjectId.Delete(ModelId(vehicle.id))

	return true
}
