package controllers

import (
	"net/http"
)

func permissionHandler(w http.ResponseWriter, r *http.Request) {
	// log.Printf("controllers/permissionsHandler.go > INFO > method: %v, %v\n", r.Method, r.RequestURI)

	// v := mux.Vars(r)                   // Read the variables in the URL.
	// id := stringToUint(fmt.Sprintf("%v", v["id"]))
	// role, err := models.RoleByLanguageAndID(language, id)
	// if err != nil {
	// 	panic(err)
	// }
	// routePermissions, err := models.RoutePermissionsByLanguageAndRoleID(language, id)
	// if err != nil {
	// 	panic(err)
	// }
	// context := views.RoutePermissionsView{
	// 	Page: views.Page{
	// 		MenuRoles: true,
	// 	},
	// 	Role:             role,
	// 	RoutePermissions: routePermissions,
	// }
	// render(w, r, langCookie, "routepermissionList.html", context)
}

func permissionUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// log.Printf("controllers/routePermission.go > INFO > RoutePermissionEditHandler() > method: %v, %v\n", r.Method, r.RequestURI)

	// v := mux.Vars(r)                   // Read the variables in the URL.
	// s := getSession(r)                 // Get session data.
	// id := stringToUint(fmt.Sprintf("%v", v["id"]))

	// routePermissions, err := models.RoutePermissionsByLanguageAndRoleID(language, id)
	// if err != nil {
	// 	panic(err)
	// }

	// if r.Method == http.MethodPost {
	// 	currentUserEmail := fmt.Sprintf("%v", s.Values["Email"])
	// 	r.ParseForm()

	// 	for i := 0; i < len(routePermissions); i++ {
	// 		cbp := fmt.Sprintf("%v", r.Form.Get("hiddenp"+strconv.Itoa(i)))
	// 		//fmt.Printf("hiddenp%v: %v\n", strconv.Itoa(i), cbp)
	// 		if cbp == "" {
	// 			routePermissions[i].Permission = false
	// 		} else {
	// 			routePermissions[i].Permission = true
	// 		}

	// 		routePermission, err := models.RoutePermissionByLanguageAndID(language, routePermissions[i].ID)
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		if routePermission.Permission != routePermissions[i].Permission {
	// 			routePermission.Permission = routePermissions[i].Permission
	// 			routePermission.UpdatedBy = currentUserEmail
	// 			routePermission.Update()
	// 		}
	// 	}
	// 	setPermissionMap()
	// 	http.Redirect(w, r, "/role/list", http.StatusSeeOther)
	// }

	// role, err := models.RoleByLanguageAndID(language, id)
	// if err != nil {
	// 	panic(err)
	// }

	// context := views.RoutePermissionsView{
	// 	Page: views.Page{
	// 		MenuRoles: true,
	// 	},
	// 	Role:             role,
	// 	RoutePermissions: routePermissions,
	// }
	// render(w, r, langCookie, "routepermissionList.html", context)
}
