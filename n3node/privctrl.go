package n3node

import (
	"../n3influx"
)

// // PrivCtrlRule :
// type PrivCtrlRule struct {
// 	Ctx  string
// 	Rule map[string]string // [path]->ctrl
// }

// //                                                 Object
// func mkPrivCtrl(dbClt *n3influx.DBClient) (rst map[string]PrivCtrlRule) {
// 	rst = make(map[string]PrivCtrlRule)
// 	rst1, rst2, rstLeft := dbClt.PairListOfSPO("ctxid", "O") // *** already the last rstLeft ***
// 	for i := 0; i < len(rst1); i++ {
// 		ID := rstLeft[i]
// 		if ID != MARKDelID {
// 			// fPln(rst1[i], rst2[i], ID)
// 			object := rst1[i]
// 			rst[object] = PrivCtrlRule{
// 				Ctx:  rst2[i],
// 				Rule: make(map[string]string),
// 			}
// 			Ps, Os := dbClt.POsByS("privctrl", ID, "", "")
// 			for i := 0; i < len(Ps); i++ {
// 				rst[object].Rule[Ps[i]] = Os[i]
// 			}
// 		}
// 	}
// 	return
// }

//                                                            object     context    path   ctrl
func mkPrivCtrl(dbClt *n3influx.DBClient) (ObjCtxPathCtrl map[string]map[string]map[string]string) {
	//                        object     context    path   ctrl
	ObjCtxPathCtrl = make(map[string]map[string]map[string]string)
	rst1, rst2, rstLeft := dbClt.PairListOfSPO("ctxid", "O") //     *** already the last rstLeft ***
	for i := 0; i < len(rst1); i++ {                         //     *** each rst1 is distinct ***
		ID := rstLeft[i]
		if ID != MARKDelID {

			// fPln(rst1[i], rst2[i], ID)
			object, ctx := rst1[i], rst2[i]

			if _, ok := ObjCtxPathCtrl[object]; !ok {
				//                                contex     path   ctrl
				ObjCtxPathCtrl[object] = make(map[string]map[string]string)
			}

			if _, ok := ObjCtxPathCtrl[object][ctx]; !ok {
				//                                     path   ctrl
				ObjCtxPathCtrl[object][ctx] = make(map[string]string)
			}

			verstrs := sSpl(dbClt.LastOBySP("privctrl-meta", ID, "V"), "-")
			vlow, vhigh := S(verstrs[0]).ToInt64(), S(verstrs[1]).ToInt64()
			fPln(vlow, vhigh)
			Ps, Os := dbClt.POsByS("privctrl", ID, "", "", vlow, vhigh)
			for i := 0; i < len(Ps); i++ {
				ObjCtxPathCtrl[object][ctx][Ps[i]] = Os[i]
			}
		}
	}
	return
}
