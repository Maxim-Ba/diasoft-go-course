package hw10programoptimization

import (
	"bytes"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	b.StopTimer()
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":` + "\n" +
		`"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79",` + "\n" +
		`"Password":"InAQJvsq","Address":"Blackbird Place 25"}` + "\n" +
		`{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":` + "\n" +
		`"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn"` + "\n" +
		`,"Address":"Fulton Hill 80"}` + "\n" +
		`{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com",` + "\n" +
		`"Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}` + "\n" +
		`{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net",` + "\n" +
		`"Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}` + "\n" +
		`{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com",` + "\n" +
		`"Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		GetDomainStat(bytes.NewBufferString(data), "com")
		b.StopTimer()
	}
}
