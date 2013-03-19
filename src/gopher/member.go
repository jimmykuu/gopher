/*
会员
*/

package gopher

import (
	"net/http"
)

func membersHandler(w http.ResponseWriter, r *http.Request) {
	c := DB.C("users")
	var newestMembers []User
	c.Find(nil).Sort("-joinedat").Limit(40).All(&newestMembers)

	membersCount, _ := c.Find(nil).Count()

	renderTemplate(w, r, "member/index.html", map[string]interface{}{
		"newestMembers": newestMembers,
		"membersCount":  membersCount,
		"active":        "members",
	})
}
