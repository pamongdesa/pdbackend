package helpdesk

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gocroot/helper/atapi"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// mendapatkan nama team helpdesk dari pesan
func GetNamaTeamFromPesan(Pesan itmodel.IteungMessage, db *mongo.Database) (team string, helpdeskslist []string, err error) {
	msg := strings.ReplaceAll(Pesan.Message, "bantuan", "")
	msg = strings.ReplaceAll(msg, "operator", "")
	msg = strings.TrimSpace(msg)
	//ambil dulu semua nama team di database
	helpdesks, err := atdb.GetAllDistinctDoc(db, bson.M{}, "team", "user")
	if err != nil {
		return
	}
	//pecah kalimat batasan spasi
	msgs := strings.Fields(msg)
	//jika nama team tidak ada atau hanya kata bantuan operator saja, maka keluarkan list nya
	if len(msgs) != 0 {
		msg = msgs[0]
	}
	//mendapatkan keyword dari kata pertama dalam kalimat masuk ke team yang mana
	for _, helpdesk := range helpdesks {
		tim := helpdesk.(string)
		if strings.EqualFold(msg, tim) {
			team = tim
			return
		}
		helpdeskslist = append(helpdeskslist, tim)
	}
	return
}

// mendapatkan scope helpdesk dari pesan
func GetScopeFromTeam(Pesan itmodel.IteungMessage, team string, db *mongo.Database) (scope string, scopeslist []string, err error) {
	msg := strings.ReplaceAll(Pesan.Message, "bantuan", "")
	msg = strings.ReplaceAll(msg, "operator", "")
	msg = strings.ReplaceAll(msg, team, "")
	msg = strings.TrimSpace(msg)
	filter := bson.M{
		"team": team,
	}
	//ambil dulu semua scope di db berdasarkan team
	scopes, err := atdb.GetAllDistinctDoc(db, filter, "scope", "user")
	if err != nil {
		return
	}
	//mendapatkan keyword masuk ke team yang mana
	for _, scp := range scopes {
		scpe := scp.(string)
		if strings.EqualFold(msg, scpe) {
			scope = scpe
			return
		}
		scopeslist = append(scopeslist, scpe)
	}
	return
}

// mendapatkan scope helpdesk dari pesan
func GetOperatorFromScopeandTeam(scope, team string, db *mongo.Database) (operator model.Userdomyikado, err error) {
	filter := bson.M{
		"scope": scope,
		"team":  team,
	}
	operator, err = atdb.GetOneLowestDoc[model.Userdomyikado](db, "user", filter, "jumlahantrian")
	if err != nil {
		return
	}
	operator.JumlahAntrian += 1
	filter = bson.M{
		"scope":       scope,
		"team":        team,
		"phonenumber": operator.PhoneNumber,
	}
	_, err = atdb.ReplaceOneDoc(db, "user", filter, operator)
	if err != nil {
		return
	}
	return
}

// handling key word, keyword :bantuan operator
func StartHelpdesk(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	//check apakah tiket dari user sudah di tutup atau belum
	user, err := atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"terlayani": bson.M{"$exists": false}, "phone": Pesan.Phone_number})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return err.Error()
		}
		//berarti tiket udah close semua
	} else { //ada tiket yang belum close
		msgstr := "User " + user.Nama + " (" + user.Phone + ")\nMeminta tolong kakak " + user.User.Name + " untuk mencarikan solusi dari masalahnya:\n" + user.Masalah + "\nSilahkan langsung kontak di nomor wa.me/" + user.Phone
		msgstr += "\n\nJika sudah teratasi mohon inputkan solusi yang sudah di berikan ke user melalui link berikut:\nwa.me/" + Profile.Phonenumber + "?text=" + user.ID.Hex() + "|+solusi+dari+operator+helpdesk+:+"
		dt := &itmodel.TextMessage{
			To:       user.User.PhoneNumber,
			IsGroup:  false,
			Messages: msgstr,
		}
		go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)
		reply = "Kakak kami hubungkan dengan operator kami yang bernama *" + user.User.Name + "* di nomor wa.me/" + user.User.PhoneNumber + "\nMohon tunggu sebentar kami akan kontak kakak melalui nomor tersebut.\n_Terima kasih_"
		return
	}
	//mendapatkan semua nama team dari db
	namateam, helpdeskslist, err := GetNamaTeamFromPesan(Pesan, db)
	if err != nil {
		return err.Error()
	}

	//suruh pilih nama team kalo tidak ada
	if namateam == "" {
		reply = "Silakan memilih regional dari operator yang anda tuju:\n"
		for i, helpdesk := range helpdeskslist {
			no := strconv.Itoa(i + 1)
			teamurl := strings.ReplaceAll(helpdesk, " ", "+")
			reply += no + ". Regional " + helpdesk + "\n" + "wa.me/" + Profile.Phonenumber + "?text=bantuan+operator+" + teamurl + "\n"
		}
		return
	}
	//suruh pilih scope dari bantuan team
	scope, scopelist, err := GetScopeFromTeam(Pesan, namateam, db)
	if err != nil {
		return err.Error()
	}
	//pilih scope jika belum
	if scope == "" {
		reply = "Silakan memilih asal provinsi anda dari regional " + namateam + " :\n"
		for i, scope := range scopelist {
			no := strconv.Itoa(i + 1)
			scurl := strings.ReplaceAll(scope, " ", "+")
			reply += no + ". " + scope + "\n" + "wa.me/" + Profile.Phonenumber + "?text=bantuan+operator+" + namateam + "+" + scurl + "\n"
		}
		return
	}
	//menuliskan pertanyaan bantuan
	user = model.Laporan{
		Scope: scope,
		Team:  namateam,
		Nama:  Pesan.Alias_name,
		Phone: Pesan.Phone_number,
	}
	_, err = atdb.InsertOneDoc(db, "helpdeskuser", user)
	if err != nil {
		return err.Error()
	}
	reply = "Silahkan kak " + Pesan.Alias_name + " mengetik pertanyaan atau bantuan yang ingin dijawab oleh operator: "

	return
}

// handling key word
func EndHelpdesk(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	msgs := strings.Split(Pesan.Message, "|")
	id := msgs[0]
	// Mengonversi id string ke primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		reply = "Invalid ID format: " + err.Error()
		return
	}
	helpdeskuser, err := atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"_id": objectID, "user.phonenumber": Pesan.Phone_number})
	if err != nil {
		reply = err.Error()
		return
	}
	helpdeskuser.Solusi = strings.Split(msgs[1], ":")[1]
	helpdeskuser.Terlayani = true
	_, err = atdb.ReplaceOneDoc(db, "helpdeskuser", bson.M{"_id": objectID}, helpdeskuser)
	if err != nil {
		reply = err.Error()
		return
	}
	op := helpdeskuser.User
	op.JumlahAntrian -= 1
	filter := bson.M{
		"scope":       op.Scope,
		"team":        op.Team,
		"phonenumber": op.PhoneNumber,
	}
	_, err = atdb.ReplaceOneDoc(db, "user", filter, op)
	if err != nil {
		reply = err.Error()
		return
	}
	reply = "*Penutupan Tiket Helpdesk*\nUser : " + helpdeskuser.Nama + "\nMasalah:\n" + helpdeskuser.Masalah

	msgstr := "*Permintaan Feedback Helpdesk*\nOperator " + helpdeskuser.User.Name + " (" + helpdeskuser.User.PhoneNumber + ")\nMeminta tolong kakak " + helpdeskuser.User.Name + " untuk memberikan rating layanan (bintang 1-5) di link berikut:\n"
	msgstr += "wa.me/" + Profile.Phonenumber + "?text=" + helpdeskuser.ID.Hex() + "|+rating+bintang+layanan+helpdesk+:+5"
	dt := &itmodel.TextMessage{
		To:       helpdeskuser.Phone,
		IsGroup:  false,
		Messages: msgstr,
	}
	go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)

	return
}

// handling key word
func FeedbackHelpdesk(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string) {
	msgs := strings.Split(Pesan.Message, "|")
	id := msgs[0]
	// Mengonversi id string ke primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		reply = "Invalid ID format: " + err.Error()
		return
	}
	helpdeskuser, err := atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"_id": objectID, "phone": Pesan.Phone_number})
	if err != nil {
		reply = err.Error()
		return
	}
	strrate := strings.Split(msgs[1], ":")[1]
	rate := strings.TrimSpace(strrate)
	rt, err := strconv.Atoi(rate)
	if err != nil {
		reply = err.Error()
		return
	}
	helpdeskuser.RateLayanan = rt
	_, err = atdb.ReplaceOneDoc(db, "helpdeskuser", bson.M{"_id": objectID}, helpdeskuser)
	if err != nil {
		reply = err.Error()
		return
	}

	reply = "Terima kasih atas pemberian feedback ke operator " + helpdeskuser.User.Name + "\nSemoga kami selalu bisa melayani lebih baik lagi."

	msgstr := "*Feedback Diterima*\nUser " + helpdeskuser.Nama + " (" + helpdeskuser.Phone + ")\nMemberikan rating " + rate + " bintang"
	dt := &itmodel.TextMessage{
		To:       helpdeskuser.User.PhoneNumber,
		IsGroup:  false,
		Messages: msgstr,
	}
	go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)

	return
}

// handling non key word
func PenugasanOperator(Profile itmodel.Profile, Pesan itmodel.IteungMessage, db *mongo.Database) (reply string, err error) {
	//check apakah tiket dari user sudah di tutup atau belum
	user, err := atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"phone": Pesan.Phone_number})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			//check apakah dia operator yang belum tutup tiketnya
			user, err = atdb.GetOneLatestDoc[model.Laporan](db, "helpdeskuser", bson.M{"terlayani": bson.M{"$exists": false}, "user.phonenumber": Pesan.Phone_number})
			if err != nil {
				if err == mongo.ErrNoDocuments {
					err = nil
					reply = ""
					return
				}
				err = errors.New("galat di collection helpdeskuser operator: " + err.Error())
				return
			}
			//jika ada tiket yang statusnya belum closed
			reply = "User " + user.Nama + " (" + user.Phone + ")\nMeminta tolong kakak " + user.User.Name + " untuk mencarikan solusi dari masalahnya:\n" + user.Masalah + "\nSilahkan langsung kontak di nomor wa.me/" + user.Phone
			reply += "\n\nJika sudah teratasi mohon inputkan solusi yang sudah di berikan ke user melalui link berikut:\nwa.me/" + Profile.Phonenumber + "?text=" + user.ID.Hex() + "|+solusi+dari+operator+helpdesk+:+"
			return

		}
		err = errors.New("galat di collection helpdeskuser user: " + err.Error())
		return
	}
	if !user.Terlayani {
		user.Masalah += "\n" + Pesan.Message
		if user.User.Name == "" || user.User.PhoneNumber == "" {
			var op model.Userdomyikado
			op, err = GetOperatorFromScopeandTeam(user.Scope, user.Team, db)
			if err != nil {
				return
			}
			user.User = op
		}
		_, err = atdb.ReplaceOneDoc(db, "helpdeskuser", bson.M{"_id": user.ID}, user)
		if err != nil {
			return
		}

		msgstr := "User " + user.Nama + " (" + user.Phone + ")\nMeminta tolong kakak " + user.User.Name + " untuk mencarikan solusi dari masalahnya:\n" + user.Masalah + "\nSilahkan langsung kontak di nomor wa.me/" + user.Phone
		msgstr += "\n\nJika sudah teratasi mohon inputkan solusi yang sudah di berikan ke user melalui link berikut:\nwa.me/" + Profile.Phonenumber + "?text=" + user.ID.Hex() + "|+solusi+dari+operator+helpdesk+:+"
		dt := &itmodel.TextMessage{
			To:       user.User.PhoneNumber,
			IsGroup:  false,
			Messages: msgstr,
		}
		go atapi.PostStructWithToken[itmodel.Response]("Token", Profile.Token, dt, Profile.URLAPIText)
		reply = "Kakak kami hubungkan dengan operator kami yang bernama *" + user.User.Name + "* di nomor wa.me/" + user.User.PhoneNumber + "\nMohon tunggu sebentar kami akan kontak kakak melalui nomor tersebut.\n_Terima kasih_"

	}
	return

}
