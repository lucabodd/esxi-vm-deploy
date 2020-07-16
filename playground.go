package main

import (

	"fmt"
	ansibler "github.com/lucabodd/go-ansible"
)

func main() {

    playbook := &ansibler.PlaybookCmd{
		Playbook: "test_site.yml",
		ConnectionOptions: &ansibler.PlaybookConnectionOptions{
			Connection: "local",
		},
		Options: &ansibler.PlaybookOptions{
			Inventory: "127.0.0.1,",
			ExtraVars: map[string]interface{}{
				"string": "testing an string",
				"bool":   true,
				"int":    10,
				"array":  []string{"one", "two"},
				"dict": map[string]bool{
					"one": true,
					"two": false,
				},
			},
		},
	}


	res, err := playbook.Run()
    check(err)
	err = res.PlaybookResultsChecks()
    check(err)
	fmt.Println(res.RawStdout)

}
func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}
