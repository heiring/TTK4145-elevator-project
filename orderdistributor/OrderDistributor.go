package orderdistributor

import (
	"fmt"
	"math"
	"time"

	"../config"
	"../elevio"
	"../tools"
)

var prioritizedOrders = make([]int, 0)

func DistributeOrders(localID string, allOrders [4][3]int, allMovementDirection, allLocations map[string]int) {
	// Don't care which elev got the order
	// Active orders should not be moved unless the elev looses connection,
	// which should happen automatically, need to use the isAlive value
	// CAB buttons should only be taken by this elev

	// Loop through all orders(!)
	// Create an allorders table
	// - OR all hall orders into the table
	// - Only consider cab orders from this elev
	// Get all current elev locations and directions

	// ****** Discuss, is this necessary? ************************************************
	// If one elevator is not alive:
	// Remove all orders except the current one, in order to take extra orders as well
	// ***********************************************************************************

	curLocalFloor := allLocations[localID]
	localMovementDirection := allMovementDirection[localID]

	for row := 0; row < 4; row++ {
		curHallUpOrder := allOrders[row][elevio.BT_HallUp]
		curHallDownOrder := allOrders[row][elevio.BT_HallDown]
		curCabOrder := allOrders[row][elevio.BT_Cab]
		orderDestination := row

		localOrderDir, _ := tools.DivCheck((orderDestination - curLocalFloor), (orderDestination - curLocalFloor))
		localDistance := orderDestination - curLocalFloor
		var allDistances = make(map[string]int)
		for ID, location := range allLocations {
			allDistances[ID] = orderDestination - location
		}

		// Hall buttons
		if (curHallUpOrder != 0 || curHallDownOrder != 0) && !tools.IntInSlice(orderDestination, prioritizedOrders) {
			fmt.Println("HALL btn pressed")
			takeOrder := false
			for ID := range allMovementDirection {
				if ID != localID {
					externalOrderDir := allLocations[ID] - orderDestination
					if localOrderDir == localMovementDirection && externalOrderDir != allMovementDirection[ID] {
						// If local only elev in same dir
						takeOrder = true
					} else if localOrderDir == localMovementDirection && externalOrderDir == allMovementDirection[ID] {
						// Else if local has shortest distance
						if localDistance <= allDistances[ID] {
							takeOrder = true
						}
					} else if localOrderDir != localMovementDirection && externalOrderDir == allMovementDirection[ID] {
						// If other elevs is in same dir && local is not
						takeOrder = false
					} else if localOrderDir != localMovementDirection && externalOrderDir != allMovementDirection[ID] {
						// If no elevs in same dir
						if localMovementDirection == elevio.MD_Stop && allMovementDirection[ID] != elevio.MD_Stop {
							takeOrder = true
						} else if localMovementDirection == elevio.MD_Stop && allMovementDirection[ID] == elevio.MD_Stop {
							// Else if local has shortest distance
							if localDistance <= allDistances[ID] {
								takeOrder = true
							}
						} else if localMovementDirection != elevio.MD_Stop && allMovementDirection[ID] == elevio.MD_Stop {
							// If other elevs in STOP
							takeOrder = false
						} else if localMovementDirection != elevio.MD_Stop && allMovementDirection[ID] != elevio.MD_Stop {
							// If no elevs in STOP
							takeOrder = true
						}
					}
				}
			}

			if takeOrder {
				if curHallDownOrder != 0 {
					addOrderToQueue(elevio.BT_HallDown, orderDestination, curLocalFloor, localMovementDirection)
				}
				if curHallUpOrder != 0 {
					addOrderToQueue(elevio.BT_HallUp, orderDestination, curLocalFloor, localMovementDirection)
				}
			}
		}

		// Cab buttons
		if curCabOrder != 0 && !tools.IntInSlice(orderDestination, prioritizedOrders) {
			addOrderToQueue(elevio.BT_Cab, orderDestination, curLocalFloor, localMovementDirection)
		}
	}

}

func addOrderToQueue(button elevio.ButtonType, orderDestination, curLocalFloor, localMovementDirection int) {
	// Bugs! Elev on the way up may stop for HallDown orders
	//
	if len(prioritizedOrders) <= 0 {
		prioritizedOrders = append(prioritizedOrders, orderDestination)
		//fmt.Println("NEW ORDER APPENDED (CAB): ", orderDestination)
	} else {
		for i, lastOrder := range prioritizedOrders {
			if !tools.IntInSlice(orderDestination, prioritizedOrders) {
				lastOrderDirection, _ := tools.DivCheck(lastOrder-curLocalFloor, int(math.Abs(float64(lastOrder-curLocalFloor))))
				curOrderDirection, _ := tools.DivCheck(orderDestination-curLocalFloor, int(math.Abs(float64(orderDestination-curLocalFloor))))

				// if lastOrder not in direction but neworder is
				if (localMovementDirection != elevio.MD_Stop) && (lastOrderDirection != localMovementDirection) && (curOrderDirection == localMovementDirection) {
					if !(button == elevio.BT_HallDown && curOrderDirection == int(elevio.MD_Up)) ||
						!(button == elevio.BT_HallUp && curOrderDirection == elevio.MD_Down) {
						prioritizedOrders = append([]int{orderDestination}, prioritizedOrders...)
						break
					}

				}

				// if both orders in same dir and neworder closer than lastOrder
				if lastOrderDirection == curOrderDirection {
					newOrderDistance := int(math.Abs(float64(orderDestination - curLocalFloor)))
					lastOrderDistance := int(math.Abs(float64(lastOrder - curLocalFloor)))
					if newOrderDistance < lastOrderDistance {
						if !(button == elevio.BT_HallDown && curOrderDirection == int(elevio.MD_Up)) ||
							!(button == elevio.BT_HallUp && curOrderDirection == elevio.MD_Down) {
							prioritizedOrders = append([]int{orderDestination}, prioritizedOrders...)
							break
						}
					}
				}

				// Give new order lowest priority
				if i == (len(prioritizedOrders) - 1) {
					prioritizedOrders = append(prioritizedOrders, orderDestination)
				}
			}
		}
	}
}

func PollOrders(receiver chan<- int) {
	var prevOrder int
	init := true
	for {
		time.Sleep(config.PollRate)
		order := GetOrderFloor()
		if (order != prevOrder && order != -1) || (order == -1 && init) {
			//fmt.Println("Sending order...")
			receiver <- order
			if init {
				init = false
			}
		}
		prevOrder = order
	}
}

func CompleteCurrentOrder() {
	//fmt.Println("REMOVED COMPLETED ORDER")
	//fmt.Println("OLD PRIO.: ", prioritizedOrders)
	prioritizedOrders = prioritizedOrders[1:]
	//fmt.Println("NEW PRIO.: ", prioritizedOrders)
}

func GetOrderFloor() int {
	if len(prioritizedOrders) > 0 {
		return prioritizedOrders[0]
	}
	return -1
}
