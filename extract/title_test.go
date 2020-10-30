package extract

import (
	"io"
	"strings"
	"testing"
)

var (
	allHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ilno fake</title>
</head>
<body>
	<div id=ilno-thread data-title=ilno data-ilno-id=/new/ >
</body>
</html>
	`
	withoutDIDHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ilno fake</title>
</head>
<body>
	<div id=ilno-thread data-title=ilno ></div>
</body>
</html>
	`
	withoutDTHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ilno</title>
</head>
<body>
	<div id=ilno-thread></div>
</body>
</html>
	`
	withoutTitleHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body>
	<div id=ilno-thread></div>
</body>
</html>
	`
	withoutRootHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body>
</body>
</html>
	`
	InvalidHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ilno</title>
</head>
<body>
	<div id=ilno-thread></div>
	<div
	<p></span>
</body>
</html>
	`
)

func TestTitleAndThreadURI(t *testing.T) {
	type args struct {
		body         io.Reader
		defaultTitle string
		defaultURI   string
	}
	tests := []struct {
		name      string
		args      args
		wantTitle string
		wantURI   string
		wantErr   bool
	}{
		{"all", args{strings.NewReader(allHTML), "Untitled", "/"}, "ilno", "/new/", false},
		{"withoutDID", args{strings.NewReader(withoutDIDHTML), "Untitled", "/"}, "ilno", "/", false},
		{"withoutDT", args{strings.NewReader(withoutDTHTML), "Untitled", "/"}, "ilno", "/", false},
		{"withoutTitle", args{strings.NewReader(withoutTitleHTML), "Untitled", "/"}, "Untitled", "/", false},
		{"withoutRoot", args{strings.NewReader(withoutRootHTML), "Untitled", "/"}, "Untitled", "/", true},
		{"Invalid", args{strings.NewReader(InvalidHTML), "Untitled", "/"}, "ilno", "/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotURI, err := titleAndThreadURI(tt.args.body, tt.args.defaultTitle, tt.args.defaultURI)
			if (err != nil) != tt.wantErr {
				t.Errorf("TitleAndThreadURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTitle != tt.wantTitle {
				t.Errorf("TitleAndThreadURI() gotTitle = %v, want %v", gotTitle, tt.wantTitle)
			}
			if gotURI != tt.wantURI {
				t.Errorf("TitleAndThreadURI() gotUri = %v, want %v", gotURI, tt.wantURI)
			}
		})
	}
}
