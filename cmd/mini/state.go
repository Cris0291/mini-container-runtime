package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
)

func decompose(data interface{}){
  typ := reflect.TypeOf(data) 
	val := reflect.ValueOf(data)

	if(val.Kind() == reflect.Struct){
		for i := 0; i < val.NumField(); i++{
			fieldVal := val.Field(i)
			fieldType := typ.Field(i)

			if(fieldVal.Kind() == reflect.Struct){
				decompose(fieldVal.Interface())
			}
			if(fieldVal.Kind() == reflect.Slice){
				for(j := 0; j < fieldVal.Len(); j++){
					element := fieldVal.
				}
			}
		}
	}
}

func state(containerID string) error {
	containerPath := filepath.Join("/run/mycontainer", containerID)
	statePath := filepath.Join(containerPath, "state.json")
	lockPath := filepath.Join(containerPath, "lock")

	fileLock, err := os.OpenFile(lockPath, syscall.O_RDWR, 0)
	if(err != nil){
		return err
	}

	err = syscall.Flock(int(fileLock.Fd()), syscall.LOCK_EX)
	if(err != nil){
		return err
	}
	defer fileLock.Close()

	file, err := os.ReadFile(statePath)
	if(err != nil){
		return err
	}

	var state ContainerState
	err = json.Unmarshal(file, &state)
	if(err != nil){
		return err
	}
	
}
