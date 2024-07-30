package app_adinfo

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"httpr2/mw_template"
	"net/http"
	"os"

	ldap "gopkg.in/ldap.v2"
)

type LDAPConfig struct {
	LDAPAddress  string //dc.example.tld:636
	BindDN       string
	BindPW       string
	BaseDN       string
	BaseFilter   string   //(objectClass=person)
	BaseControls []string //[]string{"dn", "cn", "mail", "memberOf"}
}

type User struct {
	DN         string
	Attributes map[string]string
	Groups     []string
}

func Main() *http.ServeMux {
	appRouter := http.NewServeMux()
	appRouter.HandleFunc("/", handler)
	return appRouter
}

func fetchUsers(LC LDAPConfig) ([]User, error) {
	l, err := ldap.DialTLS("tcp", LC.LDAPAddress, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	defer l.Close()

	err = l.Bind(LC.BindDN, LC.BindPW)
	if err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		LC.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		LC.BaseFilter,
		LC.BaseControls,
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var users []User
	for _, entry := range sr.Entries {
		r := map[string]string{}
		for _, u0 := range LC.BaseControls {
			if u0 != "memberOf" {
				r[u0] = entry.GetAttributeValue(u0)
			}
		}

		user := User{
			DN:         entry.DN,
			Attributes: r,
		}

		groups := entry.GetAttributeValues("memberOf")
		user.Groups = make([]string, len(groups))
		copy(user.Groups, groups)

		users = append(users, user)
	}

	return users, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	TLC := LDAPConfig{}
	LoadFile("./ldapconfig.json", &TLC)
	users, err := fetchUsers(TLC)
	if err != nil {
		//apps.Default500
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	mw_template.ProcessTemplate(w, "adminAdUsers.html", "./html-templates", 200, users)

}

func MakeDefault() {
	TLC := LDAPConfig{LDAPAddress: "dc.example.tld:636", BindDN: "", BindPW: "", BaseDN: "", BaseFilter: "(objectClass=person)", BaseControls: []string{"dn", "cn", "mail", "memberOf"}}
	SaveFile("./_ldapconfig.json", &TLC)
}

func SaveFile(xfile string, a any) {
	file, _ := os.Create(xfile)
	encoder := json.NewEncoder(file)
	err := encoder.Encode(a)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
	}
	file.Close()
	fmt.Println("Saved content to: ", xfile)
}

func LoadFile(xfile string, a any) {
	if _, err := os.Stat(xfile); os.IsNotExist(err) {
		fmt.Println("Speicherdatei nicht gefunden, neue Map wird erstellt")
	} else {
		file, _ := os.Open(xfile)
		decoder := json.NewDecoder(file)
		err := decoder.Decode(a)
		if err != nil {
			fmt.Println("Fehler in JSON-Datei:", err)
		}
		file.Close()
	}
	fmt.Println("Loaded content from: ", xfile)
}
