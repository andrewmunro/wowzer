package main

import "fmt"

func unlockLua(p *Process) {
	base := p.BaseAddress()

	p.MustWrite(base+0x1191D2, []byte{0xEB})                               // CastSpellByName
	p.MustWrite(base+0x124C76, []byte{0xEB})                               // TargetUnit
	p.MustWrite(base+0x124FD7, []byte{0x90, 0x90, 0x90, 0x90, 0x90, 0x90}) // TargetNearestEnemy
	p.MustWrite(base+0x40319C, []byte{0xEB})                               // CancelShapeshiftForm
	p.MustWrite(base+0x127F89, []byte{0xEB})
	p.MustWrite(base+0x11FF4A, []byte{0xEB})
}

func relockLua(p *Process) {
	base := p.BaseAddress()

	p.Write(base+0x1191D2, []byte{0x74})                               // CastSpellByName
	p.Write(base+0x124C76, []byte{0x74})                               // TargetUnit
	p.Write(base+0x124FD7, []byte{0x0F, 0x85, 0x9B, 0x02, 0x00, 0x00}) // TargetNearestEnemy
	p.Write(base+0x40319C, []byte{0x74})                               // CancelShapeshiftForm
	p.Write(base+0x127F89, []byte{0x74})
	p.Write(base+0x11FF4A, []byte{0x74})
}

func main() {
	p, err := NewFromName("Wow.exe")
	if err != nil {
		panic(err)
	}

	unlockLua(p)

	fmt.Println("done")
}
