package main

type DigStatus byte

const (
    DigStarted = DigStatus(0)
    DigDigging = DigStatus(1)
    DigStopped = DigStatus(2)
    DigBlockBroke = DigStatus(3)
)
