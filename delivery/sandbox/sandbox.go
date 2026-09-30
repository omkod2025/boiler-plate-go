// Package sandbox คือ delivery ของ role sandbox: รับงานทีละบรรทัด (JSON) คืนผลทีละบรรทัด
// role นี้ออกแบบให้รันใน container ที่ไม่มี network (`--network none`) และไม่มี secret
// ใช้กับงานที่เสี่ยง เช่น parse ไฟล์ที่ผู้ใช้อัปโหลด — ผู้เรียก (worker) เป็นคนเก็บผลลง DB
package sandbox

import (
	"context"
	"encoding/json"
)

// Task คืองานหนึ่งชิ้น
type Task struct {
	ID    string          `json:"id"`
	Kind  string          `json:"kind"`
	Input json.RawMessage `json:"input"`
}

// Result คือผลของงาน Error ว่าง = สำเร็จ
type Result struct {
	ID     string          `json:"id"`
	Output json.RawMessage `json:"output,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Process ประมวลผลงานหนึ่งบรรทัด — ตัวอย่างนี้ตอบ echo ให้แทนที่ด้วย parser จริงตาม Kind
func Process(_ context.Context, line []byte) []byte {
	var t Task
	var r Result
	if err := json.Unmarshal(line, &t); err != nil {
		r.Error = "invalid task"
	} else {
		r.ID = t.ID
		switch t.Kind {
		case "echo":
			r.Output = t.Input
		default:
			r.Error = "unsupported kind"
		}
	}
	out, _ := json.Marshal(r)
	return out
}
