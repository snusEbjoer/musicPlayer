// Sample Go code for user authorization

package main

import (
	"log"
	"main/youtube"
)

func main() {
	youtube := youtube.C{}
	// _, err := youtube.DownloadVideo("LJxF-i0kvPM")
	err := youtube.Download("https://rr3---sn-jvhnu5g-c35d.googlevideo.com/videoplayback?expire=1709153067&ei=y0bfZYmDMtSyv_IP-4SP0As&ip=46.138.6.8&id=o-ANzAZ682bowAa2fOg4Cz1BgayS_HvrOnd4tPn-RXiqpq&itag=243&aitags=133%2C134%2C135%2C136%2C160%2C242%2C243%2C244%2C247%2C278%2C298%2C299%2C302%2C303&source=youtube&requiressl=yes&xpc=EgVo2aDSNQ%3D%3D&mh=uJ&mm=31%2C29&mn=sn-jvhnu5g-c35d%2Csn-jvhnu5g-n8vy&ms=au%2Crdu&mv=m&mvi=3&pl=22&initcwndbps=2021250&siu=1&spc=UWF9f_IOd3Dln1LBMBDrYyxGcUHk99QX02lVcvTxviFeS1x-YH5r9oOTnQEk&vprv=1&svpuc=1&mime=video%2Fwebm&ns=gxuiTKrupS8v14yuOMing98Q&gir=yes&clen=27295235&dur=1331.133&lmt=1707003621512446&mt=1709131036&fvip=2&keepalive=yes&fexp=24007246&c=WEB&sefc=1&txp=4437434&n=cQ0XLAEJBdkzFVzQ&sparams=expire%2Cei%2Cip%2Cid%2Caitags%2Csource%2Crequiressl%2Cxpc%2Csiu%2Cspc%2Cvprv%2Csvpuc%2Cmime%2Cns%2Cgir%2Cclen%2Cdur%2Clmt&sig=AJfQdSswRAIgDdM_EuJU5MkrETiP8SOV92_XN_PSSg-yoioUIOPKlPICIGT3YWkZcOX_HLQprmO2GrQWnTM3avW-kzXYw0JFE2zb&lsparams=mh%2Cmm%2Cmn%2Cms%2Cmv%2Cmvi%2Cpl%2Cinitcwndbps&lsig=APTiJQcwRQIhAJu_DgZ-k6uC0Cm0vsoIxRXYQ6hAKLzmwmw-tg0X1NegAiB7KOlbpqWu4dlR0JekZhPh4lzyq6lgU8Si8i_nDbljlg%3D%3D")
	if err != nil {
		log.Fatal(err)
	}

}
