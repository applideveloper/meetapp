package views

import (
	"net/http"

	"time"

	"encoding/json"
	"fmt"

	"github.com/go-xweb/uuid"
	"github.com/guregu/kami"
	"github.com/shumipro/meetapp/server/models"
	"github.com/shumipro/meetapp/server/oauth"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
)

func init() {
	kami.Get("/login", Login)
	kami.Get("/logout", Logout)
	kami.Get("/login/facebook", LoginFacebook)
	kami.Get("/auth/callback", AuthCallback)
}

func Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if _, ok := oauth.FromContext(ctx); ok {
		// login済みならmypageへ
		http.Redirect(w, r, "/u/mypage", 302)
		return
	}

	preload := TemplateHeader{
		Title: "Login",
	}
	ExecuteTemplate(ctx, w, "login", preload)
}

func Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	oauth.ResetCacheAuthToken(ctx, w)
	http.Redirect(w, r, "/login", 302)
}

func LoginFacebook(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	c := oauth.Facebook(ctx)
	http.Redirect(w, r, c.AuthCodeURL(""), 302)
}

func AuthCallback(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := oauth.GetFacebookAuthToken(ctx, code)
	if err != nil {
		panic(err.Error())
	}

	facebookID, res, err := oauth.GetFacebookMe(ctx, token.AccessToken)
	if err != nil {
		panic(err.Error())
	}

	user, err := models.UsersTable().FindByFacebookID(ctx, facebookID)
	if err == mgo.ErrNotFound {
		// 新規
		userID := uuid.New()

		user = models.User{}
		user.ID = userID

		var fbUser models.FacebookUser
		data, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal(data, &fbUser); err != nil {
			panic(err)
		}
		user.Name = fbUser.Name // TODO: 一旦Facebookオンリーなので
		user.ImageURL = user.IconImageURL()
		user.FBUser = fbUser

		nowTime := time.Now()
		user.CreateAt = nowTime
		user.UpdateAt = nowTime

		// 登録する
		if err := models.UsersTable().Upsert(ctx, user); err != nil {
			panic(err)
		} else {
			fmt.Println("とうろくした")
		}
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("とうろくずみ")
	}

	// RedisでCacheとCookieに書き込む
	err = oauth.CacheAuthToken(ctx, w, user.ID, *token)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/u/mypage", 302)
}
