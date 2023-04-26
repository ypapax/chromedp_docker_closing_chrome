package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"github.com/ypapax/logrus_conf"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


const chromeDpReuse = "chromedpreuse"

func main() {
	t := os.Getenv("TYPE")
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic is caught: %+v", r)
		}
	}()
	log.Println("app is running")
	cycle := 0
	log.SetFlags(log.LstdFlags | log.Llongfile)
	//LogChromeMem()
	//ctx := context.Background()

	if err := func() error {
		if err := logrus_conf.PrepareFromEnv("headless_browser"); err != nil {
			return errors.WithStack(err)
		}
		simultStr := os.Getenv("SIMULT")
		if len(simultStr) == 0 {
			return errors.Errorf("missing simult env")
		}
		simult, err := strconv.Atoi(simultStr)
		if err != nil {
			return errors.WithStack(err)
		}
		if t == chromeDpReuse {
			generateCommonContexts(simult)
		}
		log.Printf("simult: %+v", simult)
		simultControl := make(chan time.Time, simult)
		var latestErrTime, latestOkTime time.Time
		//var count int
		var countNoErr int
		var countErr int
		var countMtx sync.Mutex
		var started = time.Now()
		parseSites()
		for {
			cycle++
			simultControl<-time.Now()
			if err := Mem(); err != nil {
				return errors.WithStack(err)
			}
			of, err := OpenFiles()
			if err != nil {
				return errors.WithStack(err)
			}
			log.Printf("open files: %+v", of)

			go func(cycle int){
				defer func(){
					if r := recover(); r != nil {
						log.Printf("panic is caught: %+v", r)
					}
				}()
				defer func(){
					<-simultControl
				}()
				if errF := func() error {
					log.Printf("starting cycle %+v\n", cycle)
					u := randomSite()
					var f func(string)(string, error)

					switch t {
					case "selenium":
						f = seleniumRunChrome
					case "seleniumff":
						f = seleniumRunFirefox
					case "chromedp":
						f = chromedpRunProperClose
					case chromeDpReuse:
						f = chromedpRunReuseContext
					default:
						return errors.Errorf("type %+v is not supported", t)
					}
					title, errF := f(u)
					if errF != nil {
						log.Printf("error for url '%+v': %+v", u, errF)
						func(){
							countMtx.Lock()
							defer countMtx.Unlock()
							countErr++
							latestErrTime = time.Now()
						}()
					} else {
						func(){
							countMtx.Lock()
							defer countMtx.Unlock()
							countNoErr++
							latestOkTime = time.Now()
						}()
					}
					func(){
						countMtx.Lock()
						defer countMtx.Unlock()
						diff := time.Since(started)
						diffMinutes := diff.Minutes()
						totalSpeedInMinuteNoErr := float64(countNoErr) / diffMinutes
						log.Printf("title: %+v, %+v stats: countErr: %+v, countNoErr: %+v, total: %+v, totalSpeedInMinuteNoErr: %+v, latest err time: %+v(%+v), latest ok time: %+v(%+v) for url %+v",
							title, diff, countErr, countNoErr, countErr + countNoErr, totalSpeedInMinuteNoErr, latestErrTime, time.Since(latestErrTime), latestOkTime, time.Since(latestOkTime), u)
					}()
					return nil
				}(); errF != nil {
					log.Printf("error is caught: %+v", errF)
				}

				/*sl := time.Second
				log.Printf("sleeping for %s\n", sl)
				time.Sleep(sl)*/
			}(cycle)

		}
	}(); err != nil {
		log.Printf("error: %+v", err)
	}

}

var (
	commonContextMtx       = &sync.Mutex{}
	//commonContextPoolInUse []context.Context
	commonContextPoolFree  []context.Context
)

func generateCommonContexts(count int) {
	commonContextMtx.Lock()
	defer commonContextMtx.Unlock()
	for i := 0; i < count; i++ {
		ctx0, _ := chromedp.NewContext(
			context.Background(),
			chromedp.WithLogf(log.Printf),
		)
		commonContextPoolFree = append(commonContextPoolFree, ctx0)
	}
}
func getFreeContext() (*context.Context, error) {
	commonContextMtx.Lock()
	defer commonContextMtx.Unlock()
	if len(commonContextPoolFree) == 0 {
		return nil, errors.Errorf("no free context")
	}
	result := commonContextPoolFree[0]
	commonContextPoolFree = commonContextPoolFree[1:]
	return &result, nil
}

func returnContextBack(ctx context.Context) {
	commonContextMtx.Lock()
	defer commonContextMtx.Unlock()
	commonContextPoolFree = append(commonContextPoolFree, ctx)
}

func chromedpRunReuseContext(u string) (string, error) {
	ctx0, err := getFreeContext()
	if err != nil {
		return "", errors.WithStack(err)
	}
	if ctx0 == nil {
		return "", errors.Errorf("ctx0 is nil")
	}
	defer func(){
		returnContextBack(*ctx0)
	}()
	selector := `title`
	log.Println("requesting", u)
	log.Println("selector", selector)
	var result string

	if errR := chromedp.Run(*ctx0,
		chromedp.Navigate(u),
		chromedp.WaitReady(selector),
		chromedp.OuterHTML(selector, &result),
	); errR != nil {
		return "", errors.WithStack(errR)
	}
	return result, nil
}


func chromedpRunProperClose(u string) (string, error) {
	ctx0, cancel2 := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel2()
	defer ctx0.Done()

	selector := `title`
	log.Println("requesting", u)
	log.Println("selector", selector)
	var result string
	err := chromedp.Run(ctx0,
		chromedp.Navigate(u),
		chromedp.WaitReady(selector),
		chromedp.OuterHTML(selector, &result),
	)
	if err != nil {
		return "", errors.WithStack(err)
	}
	log.Printf("result: %s", result)
	if errCancel := chromedp.Cancel(ctx0); errCancel != nil {
		return "", errors.WithStack(errCancel)
	} else {
		log.Printf("cancel run without an error!")
	}
	return result, nil
}



var sites []string
func parseSites() {
	sitesRaw := strings.Split(sitesLines, "\n")
	for _, s := range sitesRaw {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		sites = append(sites, "https://"+s)
	}
	rand.Seed(time.Now().Unix())
}

func randomSite() string {
	i := rand.Intn(len(sites))
	return sites[i]
}

const sitesLines = `google.com
facebook.com
youtube.com
baidu.com
yahoo.com
amazon.com
wikipedia.org
qq.com
twitter.com
slashdot.org
google.co.in
taobao.com
live.com
sina.com.cn
yahoo.co.jp
linkedin.com
weibo.com
ebay.com
google.co.jp
yandex.ru
bing.com
vk.com
hao123.com
google.de
instagram.com
t.co
msn.com
amazon.co.jp
tmall.com
google.co.uk
pinterest.com
ask.com
reddit.com
wordpress.com
mail.ru
google.fr
blogspot.com
paypal.com
onclickads.net
google.com.br
tumblr.com
apple.com
google.ru
aliexpress.com
sohu.com
microsoft.com
imgur.com
xvideos.com
google.it
imdb.com
google.es
netflix.com
gmw.cn
amazon.de
fc2.com
360.cn
alibaba.com
go.com
stackoverflow.com
ok.ru
google.com.mx
google.ca
amazon.in
google.com.hk
tianya.cn
amazon.co.uk
craigslist.org
pornhub.com
rakuten.co.jp
naver.com
blogger.com
diply.com
google.com.tr
xhamster.com
flipkart.com
espn.go.com
soso.com
outbrain.com
nicovideo.jp
google.co.id
cnn.com
xinhuanet.com
dropbox.com
google.co.kr
googleusercontent.com
github.com
bongacams.com
ebay.de
kat.cr
bbc.co.uk
google.pl
google.com.au
pixnet.net
tradeadexchange.com
popads.net
googleadservices.com
ebay.co.uk
dailymotion.com
sogou.com
adnetworkperformance.com
adobe.com
directrev.com
nytimes.com
jd.com
wikia.com
adcash.com
livedoor.jp
booking.com
163.com
bbc.com
alipay.com
coccoc.com
dailymail.co.uk
indiatimes.com
china.com
dmm.co.jp
china.com.cn
chase.com
xnxx.com
buzzfeed.com
google.com.sa
huffingtonpost.com
youku.com
google.com.eg
google.com.tw
terraclicks.com
uol.com.br
amazon.cn
snapdeal.com
office.com
google.com.ar
microsoftonline.com
walmart.com
ameblo.jp
amazon.fr
daum.net
amazonaws.com
blogspot.in
slideshare.net
etsy.com
twitch.tv
google.com.pk
whatsapp.com
bankofamerica.com
yelp.com
globo.com
theguardian.com
tudou.com
flickr.com
aol.com
stackexchange.com
chinadaily.com.cn
cnet.com
weather.com
indeed.com
ettoday.net
amazon.it
reimageplus.com
quora.com
redtube.com
soundcloud.com
detail.tmall.com
google.nl
forbes.com
douban.com
loading-delivery2.com
naver.jp
bp.blogspot.com
cntv.cn
cnzz.com
google.co.za
wellsfargo.com
google.co.ve
target.com
youporn.com
adf.ly
zillow.com
vice.com
google.gr
leboncoin.fr
kakaku.com
ikea.com
gmail.com
bestbuy.com
vimeo.com
avito.ru
godaddy.com
spaceshipads.com
goo.ne.jp
salesforce.com
about.com
tripadvisor.com
allegro.pl
livejournal.com
nih.gov
tubecup.com
adplxmd.com
foxnews.com
deviantart.com
files.wordpress.com
doublepimp.com
google.com.ua
washingtonpost.com
theladbible.com
w3schools.com
themeforest.net
feedly.com
wikihow.com
wordpress.org
office365.com
taboola.com
9gag.com
mozilla.org
akamaihd.net
zol.com.cn
hclips.com
mediafire.com
businessinsider.com
google.cn
onet.pl
comcast.net
gfycat.com
softonic.com
google.com.co
pixiv.net
google.co.th
zhihu.com
americanexpress.com
amazon.es
mystart.com
nfl.com
wix.com
steamcommunity.com
archive.org
usps.com
ups.com
google.com.sg
wikimedia.org
bilibili.com
homedepot.com
google.ro
secureserver.net
doorblog.jp
force.com
telegraph.co.uk
skype.com
detik.com
shutterstock.com
google.com.ng
ebay-kleinanzeigen.de
weebly.com
popcash.net
google.com.ph
addthis.com
steampowered.com
web.de
bitauto.com
blogspot.com.br
google.se
github.io
rambler.ru
avg.com
ndtv.com
hulu.com
gamer.com.tw
xywy.com
huanqiu.com
nametests.com
51.la
orange.fr
tlbb8.com
sourceforge.net
hdfcbank.com
livejasmin.com
espncricinfo.com
answers.com
hp.com
gmx.net
youm7.com
mailchimp.com
mercadolivre.com.br
speedtest.net
xfinity.com
ebay.in
webmd.com
ifeng.com
google.at
groupon.com
blogfa.com
wordreference.com
uptodown.com
xuite.net
media.tumblr.com
hootsuite.com
usatoday.com
google.pt
capitalone.com
stumbleupon.com
goodreads.com
wp.pl
people.com.cn
bet365.com
google.be
t-online.de
paytm.com
fedex.com
fbcdn.net
icicibank.com
blog.jp
google.com.pe
thesaurus.com
bloomberg.com
mashable.com
caijing.com.cn
bild.de
extratorrent.cc
warmportrait.com
dmm.com
pandora.com
putlocker.is
amazon.ca
spiegel.de
seznam.cz
google.ae
spotify.com
wsj.com
dell.com
ign.com
jabong.com
udn.com
2ch.net
macys.com
chaturbate.com
kaskus.co.id
att.com
engadget.com
accuweather.com
gameforge.com
varzesh3.com
watsons.tmall.com
life.com.tw
smzdm.com
badoo.com
google.ch
mama.cn
samsung.com
adidas.tmall.com
rutracker.org
1688.com
chaoshi.tmall.com
1905.com
gsmarena.com
google.az
youth.cn
onlinesbi.com
styletv.com.cn
abs-cbnnews.com
mega.nz
twimg.com
liveadexchanger.com
livedoor.biz
upornia.com
zendesk.com
trello.com
mlb.com
rediff.com
tistory.com
39.net
reference.com
google.cl
google.com.bd
google.cz
milliyet.com.tr
reuters.com
icloud.com
verizonwireless.com
haosou.com
liputan6.com
kohls.com
kickstarter.com
kouclo.com
sahibinden.com
shopclues.com
enet.com.cn
ebay.it
mydomainadvisor.com
iqiyi.com
sberbank.ru
impress.co.jp
eksisozluk.com
bleacherreport.com
slickdeals.net
yaolan.com
tube8.com
evernote.com
trackingclick.net
babytree.com
baike.com
lady8844.com
infusionsoft.com
hurriyet.com.tr
ask.fm
google.hu
liveinternet.ru
flirchi.com
newegg.com
ijreview.com
torrentz.eu
vid.me
likes.com
kinopoisk.ru
thefreedictionary.com
youradexchange.com
pinimg.com
oracle.com
ppomppu.co.kr
google.ie
gap.com
4shared.com
rt.com
google.co.il
yandex.ua
scribd.com
ebay.com.au
quikr.com
photobucket.com
ltn.com.tw
taleo.net
repubblica.it
ce.cn
libero.it
onedio.com
list-manage.com
uploaded.net
slack.com
blogspot.com.es
blogimg.jp
livedoor.com
meetup.com
cbssports.com
retailmenot.com
goal.com
goodgamestudios.com
cnnic.cn
eastday.com
citi.com
lifehacker.com
51yes.com
exoclick.com
buzzfil.net
olx.in
hm.com
neobux.com
ameba.jp
cloudfront.net
teepr.com
pconline.com.cn
google.dz
kinogo.co
gizmodo.com
elpais.com
savefrom.net
rbc.ru
disqus.com
fiverr.com
theverge.com
ewt.cc
marca.com
xda-developers.com
lowes.com
free.fr
google.fi
allrecipes.com
xe.com
battle.net
torrentz.in
kompas.com
surveymonkey.com
aparat.com
souq.com
ilividnewtab.com
mobile.de
nordstrom.com
stockstar.com
nyaa.se
time.com
asos.com
intuit.com
youboy.com
nbcnews.com
naukri.com
4dsply.com
epweike.com
streamcloud.eu
techcrunch.com
medium.com
tabelog.com
independent.co.uk
chip.de
zippyshare.com
lenovo.com
expedia.com
wunderground.com
java.com
corriere.it
gmarket.co.kr
subscene.com
webssearches.com
plarium.com
hotels.com
autohome.com.cn
playstation.com
irctc.co.in
glassdoor.com
eyny.com
ancestry.com
gamefaqs.com
sabq.org
qunar.com
myway.com
google.sk
cnbeta.com
urdupoint.com
17ok.com
albawabhnews.com
youtube-mp3.org
blackboard.com
airbnb.com
google.com.vn
hatena.ne.jp
azlyrics.com
mercadolibre.com.ar
nifty.com
ero-advertising.com
kijiji.ca
doubleclick.net
justdial.com
6pm.com
mercadolibre.com.ve
shopify.com
olx.pl
instructables.com
bestadbid.com
realtor.com
chinaz.com
costco.com
nike.com
people.com
npr.org
timeanddate.com
gmanetwork.com
issuu.com
digikala.com
lenta.ru
kayak.com
jimdo.com
subito.it
beeg.com
codecanyon.net
box.com
rottentomatoes.com
kooora.com
vcommission.com
seesaa.net
verizon.com
siteadvisor.com
discovercard.com
blogspot.jp
elmundo.es
xunlei.com
11st.co.kr
tmz.com
douyutv.com
donga.com
google.no
taringa.net
haber7.com
youdao.com
okcupid.com
bukalapak.com
clien.net
thepiratebay.la
microsoftstore.com
gazeta.pl
bhaskar.com
all2lnk.com
mirror.co.uk
hupu.com
sh.st
k618.cn
instructure.com
so-net.ne.jp
ebay.fr
zomato.com
squarespace.com
urbandictionary.com
focus.de
google.dk
zulily.com
wired.com
overstock.com
wetransfer.com
itmedia.co.jp
southwest.com
latimes.com
fidelity.com
b5m.com
list.tmall.com
csdn.net
nba.com
change.org
sakura.ne.jp
gearbest.com
drudgereport.com
freepik.com
moneycontrol.com
eonline.com
livescore.com
google.com.my
asana.com
vnexpress.net
airtel.in
duckduckgo.com
agoda.com
japanpost.jp
yandex.com.tr
r10.net
cookpad.com
yodobashi.com
rdsa2012.com
mixi.jp
unblocked.la
woot.com
ytimg.com
php.net
pof.com
makemytrip.com
udemy.com
wayfair.com
domaintools.com
statcounter.com
hespress.com
trulia.com
slate.com
asus.com
billdesk.com
sears.com
aweber.com
musicboxnewtab.com
wow.com
foodnetwork.com
pch.com
yts.to
ca.gov
constantcontact.com
bomb01.com
yandex.kz
blogspot.mx
researchgate.net
mihanblog.com
interia.pl
goo.gl
ensonhaber.com
superuser.com
lefigaro.fr
workercn.cn
gigazine.net
cnbc.com
eventbrite.com
swagbucks.com
suning.com
blog.me
staticwebdom.com
rednet.cn
yallakora.com
lazada.co.id
bookmyshow.com
popsugar.com
pcmag.com
staples.com
ero-video.net
chron.com
twoo.com
admtpmp127.com
houzz.com
youjizz.com
topfansgames.com
abril.com.br
tokopedia.com
abcnews.go.com
upwork.com
state.gov
ria.ru
asahi.com
wonderlandads.com
bitly.com
food.com
onlylady.com
tinyurl.com
lemonde.fr
ci123.com
nikkei.com
hatenablog.com
hostgator.com
academia.edu
megapopads.com
custhelp.com
redirectvoluum.com
zoho.com
alicdn.com
cdiscount.com
thewatchseries.to
bhphotovideo.com
howtogeek.com
mercadolibre.com.mx
norton.com
appledaily.com.tw
momoshop.com.tw
cbc.ca
biglobe.ne.jp
315che.com
semrush.com
shareasale.com
tnaflix.com
zappos.com
zara.com
egou.com
chinaso.com
gismeteo.ru
sciencedirect.com
nypost.com
byinmind.com
patch.com
fanli.com
nydailynews.com
littlethings.com
voc.com.cn
histats.com
faithtap.com
jcpenney.com
coursera.org
wp.com
topf1le.com
intoday.in
cbsnews.com
rightmove.co.uk
zing.vn
babycenter.com
auction.co.kr
wikiwiki.jp
backpage.com
weblio.jp
messenger.com
xbox.com
oeeee.com
marketwatch.com
windows.com
fanduel.com
clixsense.com
gemius.pl
aa.com
58.com
commentcamarche.net
nikkeibp.co.jp
olx.ua
bodybuilding.com
jrj.com.cn
mynet.com
nhk.or.jp
lequipe.fr
gawker.com
usaa.com
exblog.jp
clipconverter.cc
mi.com
videodownloadconverter.com
investing.com
mayoclinic.org
jumia.com.ng
rapidgator.net
behance.net
delta-homes.com
dropbooks.tv
e-hentai.org
ticketmaster.com
to8to.com
vetogate.com
leagueoflegends.com
mit.edu
gamespot.com
reverso.net
techradar.com
sbnation.com
liveleak.com
europa.eu
mobile01.com
ibm.com
zeroredirect1.com
himado.in
farsnews.com
wiktionary.org
vk.me
myfitnesspal.com
nhl.com
alexa.cn
kdnet.net
21cn.com
123cha.com
google.co.nz
altervista.org
cbs.com
1111.tmall.com
milanuncios.com
askmebazaar.com
gutefrage.net
www.gov.uk
atlassian.net
primewire.ag
saramin.co.kr
theatlantic.com
sozcu.com.tr
giphy.com
yoka.com
adp.com
okezone.com
olx.com.br
mystartsearch.com
junbi-tracker.com
acfun.tv
informer.com
delta.com
dianping.com
androidcentral.com
webex.com
cam4.com
t-mobile.com
cracked.com
usbank.com
stamplive.com
ted.com
google.bg
drtuber.com
gyazo.com
investopedia.com
adschemist.com
drom.ru
anitube.se
cityadspix.com
thesportbible.com
agar.io
zone-telechargement.com
united.com
hubspot.com
4chan.org
emol.com
netteller.com
jin115.com
infoseek.co.jp
ehow.com
uniqlo.com
geocities.jp
disq.us
mega.co.nz
leadzupc.com
atwiki.jp
sfgate.com
cpasbien.pw
novinky.cz
filehippo.com
dreamstime.com
cisco.com
getpocket.com
facenama.com
feclik.com
iflscience.com
priceline.com
blogspot.com.tr
wav.tv
thekitchn.com
gazzetta.it
thedailybeast.com
trklnks.com
pureadexchange.com
digitaltrends.com
sapo.pt
tomshardware.com
quizlet.com
android.com
rev2pub.com
mackeeper.com
nbcsports.com
123rf.com
ruten.com.tw
friv.com
yomiuri.co.jp
indianexpress.com
gamepedia.com
prpops.com
intel.com
deezer.com
deviantart.net
bluehost.com
qvc.com
tutorialspoint.com
allocine.fr
offpageads.com
eastmoney.com
india.com
india-mmm.net
roblox.com
fishki.net
adme.ru
genius.com
ew.com
vodlocker.com
seasonvar.ru
tsite.jp
ebates.com
audible.com
rakuten.ne.jp
almasryalyoum.com
welt.de
google.rs
google.by
thehindu.com
yadi.sk
nasa.gov
mapquest.com
idnes.cz
news.com.au
tagged.com
dafont.com
4pda.ru
ebay.ca
banggood.com
etao.com
chexun.com
sex.com
google.lk
haberturk.com
yandex.by
digg.com
carview.co.jp
cqnews.net
prezi.com
match.com
ampclicks.com
as.com
ultimate-guitar.com
ccm.net
topix.com
admtpmp124.com
usmagazine.com
monster.com
lapatilla.com
admtpmp123.com
fandango.com
indianrail.gov.in
coupons.com
189.cn
2chblog.jp
sky.com
yaplakal.com
yellowpages.com
wiley.com
rarbg.to
whitepages.com
mysmartprice.com
entrepreneur.com
google.com.kw
id.net
istockphoto.com
ea.com
walgreens.com
toysrus.com
mynavi.jp
noaa.gov
sankei.com
cnblogs.com
immobilienscout24.de
mackolik.com
pagesjaunes.fr
weather.gov
kotaku.com
google.hr
basecamp.com
clarin.com
quanjing.com
folha.uol.com.br
lolesports.com
otto.de
nairaland.com
lun.com
fanpage.gr
bt.com
forever21.com
vine.co
marriott.com
pchome.com.tw
padsdel.com
thisav.com`