package can

import (
    "testing"
    "bytes"
    "time"
)

func TestReadWrite(t *testing.T) {
    cb1, err := NewCANBus("vcan0")
    if err != nil {
	t.Fatal(err)
    }
    cb2, err := NewCANBus("vcan0")
    if err != nil {
	t.Fatal(err)
    }
    done := make(chan struct{})
    dataIn := []byte("test")
    go func() {
	cf := CANFrame{}
	cb1.Read(&cf)
	if bytes.Compare(dataIn, cf.Data[:cf.Len]) != 0 {
		t.Errorf("HAVE: %#x WANT: %#x", cf.Data[:cf.Len], dataIn)
	}
	done<-struct{}{}
    }()
    time.Sleep(time.Second/100)
    cf := CANFrame{}
    cf.SetAddr(0x123)
    cf.Len = uint32(len(dataIn))
    copy(cf.Data[:], dataIn)
    cb2.Write(&cf)
    <-done
}
