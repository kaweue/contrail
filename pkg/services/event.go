package services

import (
	"context"
	fmt "fmt"

	"github.com/pkg/errors"
)

const (
	//OperationCreate for create operation.
	OperationCreate = "CREATE"
	//OperationUpdate for update operation.
	OperationUpdate = "UPDATE"
	//OperationDelete for delete operation.
	OperationDelete = "DELETE"
)

// EventOption contains options for Event.
type EventOption struct {
	UUID      string
	Operation string
	Kind      string
	Data      map[string]interface{}
}

// HasResource defines methods that might be implemented by Event.
type HasResource interface {
	GetResource() Resource
	Operation() string
}

// CanProcessService is interface for process service.
type CanProcessService interface {
	Process(ctx context.Context, service Service) (*Event, error)
}

// Resource is a generic resource interface.
type Resource interface {
	GetUUID() string
	GetParentUUID() string
	Kind() string
	// Depends Returns UUIDs of children and back references
	Depends() []string
	ToMap() map[string]interface{}
	// AddDependency adds child/backref to model
	AddDependency(i interface{})
	// RemoveDependency removes child/backref from model
	RemoveDependency(i interface{})
}

// EventList has multiple rest requests.
type EventList struct {
	Events []*Event `json:"resources" yaml:"resources"`
}

type state int

const (
	notVisited state = iota
	visited
	temporaryVisited
)

//reorder request using Tarjan's algorithm
func visitResource(uuid string, sorted []*Event,
	eventMap map[string]*Event, stateGraph map[string]state,
) (sortedList []*Event, err error) {
	if stateGraph[uuid] == temporaryVisited {
		return nil, errors.New("dependency loop found in sync request")
	}
	if stateGraph[uuid] == visited {
		return sorted, nil
	}
	stateGraph[uuid] = temporaryVisited
	event, found := eventMap[uuid]
	if !found {
		return nil, fmt.Errorf("Resource with uuid: %s not found in eventMap", uuid)
	}
	depends := event.GetResource().Depends()
	for _, refUUID := range depends {
		sorted, err = visitResource(refUUID, sorted, eventMap, stateGraph)
		if err != nil {
			return nil, err
		}
		break
	}
	stateGraph[uuid] = visited
	sorted = append(sorted, event)
	return sorted, nil
}

// Sort sorts Events by dependency using Tarjan's algorithm.
// TODO: support parent-child relationship while checking dependencies.
func (e *EventList) Sort() (err error) {
	var sorted []*Event
	stateGraph := map[string]state{}
	eventMap := map[string]*Event{}
	for _, event := range e.Events {
		uuid := event.GetResource().GetUUID()
		stateGraph[uuid] = notVisited
		eventMap[uuid] = event
	}
	foundNotVisited := true
	for foundNotVisited {
		foundNotVisited = false
		for _, event := range e.Events {
			uuid := event.GetResource().GetUUID()
			state := stateGraph[uuid]
			if state == notVisited {
				sorted, err = visitResource(uuid, sorted, eventMap, stateGraph)
				if err != nil {
					return err
				}
				foundNotVisited = true
				break
			}
		}
	}
	e.Events = sorted
	return nil
}

// Process dispatches resource event to call corresponding service functions.
func (e *Event) Process(ctx context.Context, service Service) (*Event, error) {
	return e.Request.(CanProcessService).Process(ctx, service)
}

// Process process list of events.
func (e *EventList) Process(ctx context.Context, service Service) (*EventList, error) {
	var responses []*Event
	for _, event := range e.Events {
		response, err := event.Process(ctx, service)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return &EventList{
		Events: responses,
	}, nil
}

// GetResource returns event on resource.
func (e *Event) GetResource() Resource {
	if e == nil {
		return nil
	}
	resourceEvent, ok := e.Request.(HasResource)
	if !ok {
		return nil
	}
	return resourceEvent.GetResource()
}

// Operation returns operation type.
func (e *Event) Operation() string {
	if e == nil {
		return ""
	}
	resourceEvent, ok := e.Request.(HasResource)
	if !ok {
		return ""
	}
	return resourceEvent.Operation()
}
