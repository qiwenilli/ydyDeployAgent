package main

import (
    "crypto/md5"
    "fmt"
    "io"
    // "io/ioutil"
    // "log"
    "net/http"
    "net/http/httputil"
    "strings"
    "time"
    "os"
    "os/exec"
    "errors"
    "flag"
)

const (
    upload_path string = "./abc"
    server_port string = "8090"
)

func defaultHandle(w http.ResponseWriter, r *http.Request) {
    //http.NotFoundHandler()

    //http.Error(w, "403", 403)

    http.Redirect(w, r, "http://www.jindanlicai.com", 302)

    return

    w.Header().Set("Server", "nginx 7.0")
    w.Header().Set("location", "http://www.jindanlicai.com")
    io.WriteString(w, "Welcome!")
}

func postCreate(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        printRequest(w, r, true)
    }
}

//上传
func uploadHandle(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {

        //http.Redirect(w, r, "http://www.jindanlicai.com", 302)
        http.Error(w, "403", 403)
        return


        printRequest(w, r, true)
        io.WriteString(w,
        "<html>"+
        "<form action='' method=\"post\" enctype=\"multipart/form-data\">"+
        "<input type=\"file\" name='fff'/><input type=\"submit\" value=\"Upload\"/>"+
        "</form>"+
        "</html>")

    } else if r.Method == "OPTIONS" {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "content-type")
        w.Header().Set("Access-Control-Max-Age", "30")

    } else if r.Method == "POST" {
        // ip := strings.Split(r.RemoteAddr, ":")[0]
        printRequest(w, r, false)
        msg, code, err := SaveFileFromRequest(w, r, upload_path)
        
        if err != nil {
            if code <= 0 {
                code = http.StatusInternalServerError
            }
            Error(w, msg, code, err)
        }else{
            Error(w, "ok:"+msg, http.StatusOK, nil)
        }
        return

        temp_path := upload_path + "/" + msg
        //md5
        id, _ := FileHashMD5(temp_path)
        real_path := upload_path + "/" + id
        //check exists
        if Exist(real_path) {
            Error(w, id, http.StatusCreated, nil)
            return

        }
        //response
        w.Header().Set("Access-Control-Allow-Origin", "*")
        io.WriteString(w, msg)
        fmt.Println("   upload success " + id)
        //rename
        //err = os.Rename(temp_path, real_path)
        //if err != nil {
        //    fmt.Println("ERROR TO RENAME: ", err)
        //}
    }
}


//回滚
func rollback(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        http.Error(w, "403", 403)
        return

    } else if r.Method == "OPTIONS" {
        //w.Header().Set("Access-Control-Allow-Origin", "*")
        //w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
        //w.Header().Set("Access-Control-Allow-Headers", "content-type")
        //w.Header().Set("Access-Control-Max-Age", "30")

    } else if r.Method == "POST" {
        printRequest(w, r, false)

        tag_name := r.FormValue("tag_name")
        unzip_to_path := r.FormValue("path")

        //
        temp_path := upload_path + "/" + tag_name 

        //
        if len(tag_name)<=1 {
            http.Error(w, "err:tag_name is nil", http.StatusInternalServerError)
        }else if Exist(unzip_to_path)==false{
            http.Error(w, "err:server config path error", http.StatusInternalServerError)
        }else if Exist(temp_path)==false{
            http.Error(w, "err:backup tag not exist", http.StatusInternalServerError)
        }else{
        
            //在代码上线之前执行的钩子


            //上传成功后，直接解压到工程目录
            out,err := exec.Command( `tar`,`-zxvf`, temp_path, `-C`, unzip_to_path ).Output()
            fmt.Println("---------->>>", string(out), err, temp_path, unzip_to_path)
            if err!=nil {
                http.Error(w, "err:unzip error", http.StatusInternalServerError)
                return
            }

            //在代码上线之后执行的钩子
            if Exist(z_hook)==true {
                exec.Command( `/bin/bash`, z_hook ).Output()
            }else{
                fmt.Println("no z_hook")
            }
        }
    }
}



/*--------------UTILS-------------*/

func SaveFileFromRequest(w http.ResponseWriter, r *http.Request, parent string) (string, int, error) {

    unzip_to_path := r.FormValue("path")

    fmt.Println("unzip to path",  unzip_to_path)


    fmt.Println("Reading")
    //get file
    file, head, err := r.FormFile("fff")
    if err != nil {
        return "Fail to read file from form", http.StatusInternalServerError, err

    }
    defer file.Close()

    ext_name := head.Filename[len(head.Filename)-6:len(head.Filename)]
    if ext_name!= "tar.gz"{
        return "error", http.StatusInternalServerError, errors.New("file type error") 
    }

    //temp file name
    id := fmt.Sprintf("%x", md5.Sum([]byte(head.Filename)))
    id = id + "-" + head.Filename
    temp_path := parent + "/" + id

    fmt.Println("Creating", id)
    //create file
    fW, err := os.Create(temp_path)
    if err != nil {
        return "Fail to create file!", http.StatusInternalServerError, err

    }
    defer fW.Close()

    fmt.Println("Coping")
    //save file
    _, err = io.Copy(fW, file)
    if err != nil {
        return "Fail to save file!", http.StatusInternalServerError, err
    }

    //在代码上线之前执行的钩子


    //判断布署的项目目录是否存在
    if Exist(unzip_to_path)==false {
        return "dir not exist", http.StatusInternalServerError, errors.New("project dir no exist")
    }

    //上传成功后，直接解压到工程目录
    out,err := exec.Command( `tar`,`-zxvf`, temp_path, `-C`, unzip_to_path ).Output()
    fmt.Println("---------->>>", string(out), err, temp_path, unzip_to_path)
    if err!=nil {
        return "err:unzip error", http.StatusInternalServerError, err
    }

    //在代码上线之后执行的钩子

    if Exist(z_hook)==true {
        exec.Command( `/bin/bash`, z_hook ).Output()
    }else{
        fmt.Println("no z_hook")
    }


    //return string(out), http.StatusOK, nil
    return id, http.StatusOK, nil
}

func printRequest(w http.ResponseWriter, r *http.Request, body bool) {
    fmt.Println()
    fmt.Println("------printRequest------")
    fmt.Println("requester:    " + strings.Split(r.RemoteAddr, ":")[0])
    debug(httputil.DumpRequest(r, body))
    fmt.Println("----------END-----------")
}

func Error(w http.ResponseWriter, msg string, code int, err error) {
    if err != nil {
        fmt.Println(err)

    }
    if msg != "" {
        fmt.Println(msg)

    }
    http.Error(w, msg, code)

}

// func FileHashMD5(file *os.File) (string, error) {
func FileHashMD5(path string) (string, error) {
    file, err := os.Open(path)
    defer file.Close()
    if err != nil {
        return "", err

    }
    h := md5.New()
    io.Copy(h, file)
    return fmt.Sprintf("%x", h.Sum(nil)), nil

}

func Exist(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil || os.IsExist(err)
}

func debug(data []byte, err error) {
    if err == nil {
        fmt.Printf("%s", data)

    } else {
        fmt.Printf("%s", err)

    }

}


var a_hook,z_hook string

func main() {

    _a := flag.String("a_hook", "", "exec base shell on rsync project code before")
    _z := flag.String("z_hook", "", "exec base shell on rsync project code end")
    _h := flag.Bool("h", false, "help")
    flag.Parse()

    if *_h == true {
        flag.Usage()
        os.Exit(0)
    }

    a_hook = *_a
    z_hook = *_z


    fmt.Println("starting server!")
    fmt.Println("https://127.0.0.1:" + server_port)

    // http.Post("/media", uploadHandle)
    http.HandleFunc("/", defaultHandle)
    http.HandleFunc("/media", uploadHandle)
    http.HandleFunc("/post", postCreate)
    http.HandleFunc("/back", rollback)

    //
    server := &http.Server{
        Addr: ":" + server_port,
        // Handler:        handler,
        ReadTimeout:    60 * time.Second,
        WriteTimeout:   60 * time.Second,
        MaxHeaderBytes: 1 << 10,

    }
    //err := server.ListenAndServe()
    err := server.ListenAndServeTLS("./jd.crt", "./jd.key")
    if err != nil {
        fmt.Println(err)
        return
    }
}


