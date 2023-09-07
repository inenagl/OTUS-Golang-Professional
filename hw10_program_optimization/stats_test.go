//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}

func TestGetDomainStatBorderConditions(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find empty", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	data = `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}`
	t.Run("find when there is no email field in data", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	data = `

`
	t.Run("find in empty multiline data", func(t *testing.T) {
		_, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.Error(t, err)
		require.Equal(t, "error: invalid json: ''", err.Error())
	})

	data = `{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com",Phone:"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}`
	t.Run("find in invalid json", func(t *testing.T) {
		_, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.Error(t, err)
		require.Equal(t, fmt.Sprintf("error: invalid json: '%s'", data), err.Error())
	})
}

func BenchmarkGetDomainStat(b *testing.B) {
	b.StopTimer()

	r, err := zip.OpenReader("testdata/users.dat.zip")
	require.NoError(b, err)
	defer r.Close()

	require.Equal(b, 1, len(r.File))

	for i := 0; i < b.N; i++ {
		data, err := r.File[0].Open()
		require.NoError(b, err)

		b.StartTimer()
		_, err = GetDomainStat(data, "biz")
		b.StopTimer()
		require.NoError(b, err)

		err = data.Close()
		require.NoError(b, err)
	}
}
