package main

import (
    "fmt"
    "github.com/stretchr/gomniauth"
    "github.com/stretchr/objx"
    "log"
    "net/http"
    "strings"
)

type authHandler struct {
    next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if _, err := r.Cookie("auth"); err == http.ErrNoCookie {
        w.Header().Set("Location", "/login")
        w.WriteHeader(http.StatusTemporaryRedirect)
    } else if err != nil {
        panic(err.Error())
    } else {
        h.next.ServeHTTP(w, r)
    }
}

func MustAuth(handler http.Handler) http.Handler {
    return &authHandler{next: handler}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    errWhen := "Error when trying to "
    segs := strings.Split(r.URL.Path, "/")
    if len(segs) < 3 {
        segs = strings.Split("auth/login/unknown", "/")
    }
    log.Println("segs", segs)
    action := segs[2]
    provider := segs[3]
    switch action {
    case "login":
        provider, err := gomniauth.Provider(provider)
        if err != nil {
            log.Fatalln(errWhen, "get provider", provider, "-", err)
        }
        loginUrl, err := provider.GetBeginAuthURL(nil, nil)
        if err != nil {
            log.Fatalln(errWhen, "GetBeginAuthURL for", provider, "-", err)
        }
        w.Header().Set("Location", loginUrl)
        w.WriteHeader(http.StatusTemporaryRedirect)
    case "callback":
        provider, err := gomniauth.Provider(provider)
        if err != nil {
            log.Fatalln(errWhen, "get provider", provider, "-", err)
        }
        creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
        if err != nil {
            log.Fatalln(errWhen, "complete auth for", provider, "-", err)
        }
        user, err := provider.GetUser(creds)
        if err != nil {
            log.Fatalln(errWhen, "get user from", provider, "-", err)
        }
        authCookieValue := objx.New(map[string]interface{}{
            "name": user.Name(),
        }).MustBase64()
        http.SetCookie(w, &http.Cookie{
            Name: "auth",
            Value: authCookieValue,
            Path: "/"})
        w.Header()["Location"] = []string{"/chat"}
        w.WriteHeader(http.StatusTemporaryRedirect)
    default:
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "Auth action %s not supported", action)
    }
}
