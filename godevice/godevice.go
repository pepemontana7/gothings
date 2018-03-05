package godevice

import (
	"errors"
	"fmt"
	"sync"
)

// Devices represents a customer document we
// store in our database.
type GoDevice struct {
	ID      int
	Name    string
	Status  string
	Actions []string
	Reading int
}

// db represents our internal database system.
var db = struct {
	devices map[int]GoDevice
	maxID   int
	lock    sync.Mutex
}{
	devices: map[int]GoDevice{},
}

// Initalize the database with some values.
func init() {
	Save(GoDevice{Name: "arduino-flame", Status: "active", Reading: 30, Actions: []string{"shoot"}})
	Save(GoDevice{Name: "amd-temp1", Status: "active", Reading: 40})
	Save(GoDevice{Name: "amd-temp2", Status: "active", Reading: 50})
	Save(GoDevice{Name: "amd-temp3", Status: "active", Reading: 60})
	Save(GoDevice{Name: "amd-temp4", Status: "active", Reading: 70})
	Save(GoDevice{Name: "amd-temp5", Status: "active", Reading: 80})
	Save(GoDevice{Name: "amd-temp6", Status: "active", Reading: 90})

}

// Save stores a device document in the database.
func Save(d GoDevice) (int, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	// If this customer id is out of the range
	// then we have an integrity issue.
	if d.ID > db.maxID {
		return 0, errors.New("Invalid device id")
	}

	// If no id is provided this is a new device.
	// Generate a new id.
	if d.ID == 0 {
		d.ID = db.maxID + 1
		db.maxID = d.ID
	}

	// Save the device in the database.
	db.devices[d.ID] = d

	// Return the device id.
	return d.ID, nil
}

// Update updates the customer in the database.
func Update(d GoDevice) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if _, ok := db.devices[d.ID]; !ok {
		return fmt.Errorf("device with ID %d does not exist", d.ID)
	}

	db.devices[d.ID] = d
	return nil
}

// Delete removes the customer from the database.
func Delete(d GoDevice) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if _, ok := db.devices[d.ID]; !ok {
		return fmt.Errorf("device with ID %d does not exist", d.ID)
	}
	delete(db.devices, d.ID)
	return nil
}

// Find locates a customer by id in the database.
func Find(id int) (GoDevice, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	// Locate the customer in the database.
	d, found := db.devices[id]
	if !found {
		return GoDevice{}, fmt.Errorf("device with ID %d does not exist", id)
	}

	return d, nil
}

// All returns the full database of devices.
func All() []GoDevice {
	db.lock.Lock()
	defer db.lock.Unlock()

	all := []GoDevice{}

	// Range over the map storing each device
	// in their ordered position.
	for i := 1; i <= db.maxID; i++ {
		if d, ok := db.devices[i]; ok {
			all = append(all, d)
		}
	}

	return all
}
