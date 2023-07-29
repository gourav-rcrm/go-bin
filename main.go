package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

const API_TOKEN_TEMP = "u+nsy4W8-46ea6MjzCHA"

func login() {
	cmd := exec.Command("sh", "-c", "$(curl -sL https://raw.githubusercontent.com/martindstone/pagerduty-cli/master/install.sh)")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	cmd = exec.Command("pd", "auth:set", "--token="+API_TOKEN_TEMP)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func getCommandOutput(command string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", command)
	return cmd.Output()
}

func get_user_ids() [][]interface{} {
	output, err := getCommandOutput("pd user list --output=json")
	if err != nil {
		fmt.Println("Error executing 'pd user list' command:", err)
		return nil
	}

	var data []map[string]interface{}
	err = json.Unmarshal(output, &data)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	var result_list [][]interface{}
	for _, user := range data {
		result_list = append(result_list, []interface{}{user["id"], user["summary"], user["email"]})
	}

	return result_list
}

func present(ans interface{}, users [][]interface{}) bool {
	for _, user := range users {
		if user[0] == ans {
			return true
		}
	}
	return false
}

func create_user_notification_rule(userid, contactid string, users [][]interface{}) {
	if present(userid, users) {
		cmd := exec.Command("pd", "rest", "get", "-e", fmt.Sprintf("/users/%s/contact_methods", userid), "-P", "type=phone_contact_method")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error executing 'pd rest get' command:", err)
			return
		}

		var data map[string]interface{}
		err = json.Unmarshal(output, &data)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		var all_ids [][]interface{}
		contactMethods, _ := data["contact_methods"].([]interface{})
		for _, cm := range contactMethods {
			contactMethod, _ := cm.(map[string]interface{})
			all_ids = append(all_ids, []interface{}{contactMethod["id"], contactMethod["label"], contactMethod["address"]})
		}

		conn := http.Client{}
		payload := map[string]interface{}{
			"notification_rule": map[string]interface{}{
				"type":                 "assignment_notification_rule",
				"start_delay_in_minutes": 0,
				"contact_method": map[string]interface{}{
					"id":   contactid,
					"type": "phone_contact_method",
				},
				"urgency": "high",
			},
		}
		jsonPayload, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", fmt.Sprintf("https://api.pagerduty.com/users/%s/notification_rules", userid), bytes.NewBuffer(jsonPayload))
		if err != nil {
			fmt.Println("Error creating HTTP request:", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
		req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", API_TOKEN_TEMP))

		res, err := conn.Do(req)
		if err != nil {
			fmt.Println("Error making HTTP request:", err)
			return
		}
		defer res.Body.Close()

		var responseData map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&responseData)
		if err != nil {
			fmt.Println("Error parsing HTTP response:", err)
			return
		}

		responseJSON, _ := json.MarshalIndent(responseData, "", "  ")
		fmt.Println(string(responseJSON))
	}
}

func main() {
	login()

	users := get_user_ids()
	create_user_notification_rule("P97F9YP", "PKXBF1T", users)
}

