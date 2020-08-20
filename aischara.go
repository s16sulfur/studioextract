package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sulfur/bbio"
	"github.com/vmihailenco/msgpack/v5"
)

// AISHeaderInfo strcture
type AISHeaderInfo struct {
	name    string
	version string
	pos     int64
	size    int64
}

func (info *AISHeaderInfo) fromMap(m map[string]interface{}) {
	info.name = fmt.Sprintf("%s", m["name"])
	info.version = fmt.Sprintf("%s", m["version"])
	info.pos = bbio.Cast.Int64(m["pos"])
	info.size = bbio.Cast.Int64(m["size"])
}

// AISCharaCard strcture
type AISCharaCard struct {
	startOffset    int64
	gameType       string
	sex            int32
	fullname       string
	pngSize        int64
	loadProductNo  int32
	marker         string
	loadVersion    string
	language       int32
	userID         string
	dataID         string
	infoHeaderSize int32
	infoHeader     struct {
		lstInfo []AISHeaderInfo
	}
	dataSize int64
	data     map[string][]byte
}

func (sf *AISCharaCard) findInfo(name string) (info AISHeaderInfo) {
	for _, v := range sf.infoHeader.lstInfo {
		if v.name == name {
			info = v
			return
		}
	}
	return
}

func (sf *AISCharaCard) loadPreviewInfo() (err error) {
	// Game Type
	pInfo2 := sf.findInfo("Parameter2")
	gInfo2 := sf.findInfo("GameInfo2")

	if pInfo2.name == "Parameter2" && gInfo2.name == "GameInfo2" {
		sf.gameType = "HS2"
	}
	sf.gameType = "AIS"

	// sex, name
	paraData := sf.data["Parameter"]
	para := map[string]interface{}{}
	paraErr := msgpack.Unmarshal(paraData, &para)
	if paraErr != nil {
		err = paraErr
		return
	}
	sf.sex = bbio.Cast.Int32(para["sex"])
	sf.fullname = fmt.Sprintf("%s", para["fullname"])

	return
}

// AISSceneCard strcture
type AISSceneCard struct {
	pngSize    int64
	charaCards map[string]AISCharaCard
}

var neoV2Mark = "【StudioNEOV2】"
var aisCharaMark = "【AIS_Chara】"

// AISChara strcture
type AISChara struct {
	card *AISSceneCard
}

// NewAISChara implements for AISChara
func NewAISChara() *AISChara {
	c := &AISSceneCard{}
	return &AISChara{card: c}
}

// IsNeoV2SceneCard implements for neo v2 scene
func IsNeoV2SceneCard(reader *bbio.Reader) bool {
	return (reader.Index([]byte(neoV2Mark)) > 0)
}

// GenerateFileName implements for KKChara
func (sf *AISChara) GenerateFileName(prefix string, sex int32) string {
	time.Sleep(2 * time.Millisecond)

	currentTime := time.Now()
	var fileName string
	fileName = prefix
	if sex == 0 {
		fileName += "ChaM_"
	} else if sex == 1 {
		fileName += "ChaF_"
	} else {
		fileName += "Cha_"
	}
	fileName += strings.ReplaceAll(currentTime.Format("2006.01.02.15.04.05.000"), ".", "")

	return fileName
}

// ReadChara implements for AISChara
func (sf *AISChara) ReadChara(reader *bbio.Reader, offset int64) (card AISCharaCard, err error) {
	startOffset, seekErr := reader.Seek(offset, io.SeekStart)
	if seekErr != nil {
		err = seekErr
		return
	}
	card.startOffset = int64(startOffset)

	lpno, lpnoErr := reader.ReadInt32()
	if lpnoErr != nil {
		err = lpnoErr
		return
	}
	if lpno > 100 {
		err = errors.New("Version not supported")
		return
	}
	card.loadProductNo = lpno

	mark, markErr := reader.ReadString()
	if markErr != nil {
		err = markErr
		return
	}
	if mark != aisCharaMark {
		err = errors.New("AIS Chara mark not found")
		return
	}
	card.marker = mark

	lver, lverErr := reader.ReadString()
	if lverErr != nil {
		err = lverErr
		return
	}
	card.loadVersion = lver

	lang, langErr := reader.ReadInt32()
	if langErr != nil {
		err = langErr
		return
	}
	card.language = lang

	userID, uidErr := reader.ReadString()
	if uidErr != nil {
		err = uidErr
		return
	}
	card.userID = userID

	dataID, didErr := reader.ReadString()
	if didErr != nil {
		err = didErr
		return
	}
	card.dataID = dataID

	headersz, headerszErr := reader.ReadInt32()
	if headerszErr != nil {
		err = headerszErr
		return
	}

	headerBytes := make([]byte, headersz)
	_, hrErr := reader.Read(headerBytes)
	if hrErr != nil {
		err = hrErr
		return
	}
	card.infoHeaderSize = headersz

	blockHead := map[string][]map[string]interface{}{}
	bhErr := msgpack.Unmarshal(headerBytes, &blockHead)
	if bhErr != nil {
		err = bhErr
		return
	}

	datasz, dataszErr := reader.ReadInt64()
	if dataszErr != nil {
		err = dataszErr
		return
	}
	card.dataSize = datasz

	dataOffset := reader.Position()
	card.data = make(map[string][]byte)

	infoCount := len(blockHead["lstInfo"])
	card.infoHeader.lstInfo = make([]AISHeaderInfo, infoCount)

	for i := 0; i < infoCount; i++ {
		card.infoHeader.lstInfo[i].fromMap(blockHead["lstInfo"][i])
		info := card.infoHeader.lstInfo[i]

		sbOffset := dataOffset + info.pos
		_, sbErr := reader.Seek(sbOffset, io.SeekStart)
		if sbErr != nil {
			err = sbErr
			return
		}

		bBytes := make([]byte, info.size)
		_, rbErr := reader.Read(bBytes)
		if rbErr != nil {
			err = rbErr
			return
		}

		card.data[info.name] = bBytes
	}

	loadErr := card.loadPreviewInfo()
	if loadErr != nil {
		err = loadErr
		return
	}

	return
}

// ReadScene implements for AISChara
func (sf *AISChara) ReadScene(reader *bbio.Reader, pngSize int64) (bool, error) {
	_, seekErr := reader.Seek(pngSize, io.SeekStart)
	if seekErr != nil {
		return false, seekErr
	}

	sf.card.pngSize = pngSize
	sf.card.charaCards = make(map[string]AISCharaCard)

	idxs := reader.FindAll([]byte(aisCharaMark))
	for _, v := range idxs {
		chara, cerr := sf.ReadChara(reader, int64(v-5))
		if cerr != nil {
			if isDebug {
				printError(cerr)
			} else {
				printError(errors.New("Chara card read error"))
			}

		} else {
			charFileName := sf.GenerateFileName(chara.gameType, chara.sex)
			sf.card.charaCards[charFileName] = chara
		}
	}

	return len(sf.card.charaCards) > 0, nil
}

// WriteChara implements for AISChara
func (sf *AISChara) WriteChara(card AISCharaCard, w io.Writer) (re bool, err error) {
	re = false
	writer := bbio.NewWriter(w)

	pngBytes, pngErr := createPng(252, 352, int(card.sex))
	if pngErr != nil {
		err = pngErr
		return
	}

	_, pwErr := writer.Write(pngBytes)
	if pwErr != nil {
		err = pwErr
		return
	}

	pnErr := writer.WriteInt(card.loadProductNo)
	if pnErr != nil {
		err = pnErr
		return
	}

	_, mErr := writer.WriteString(card.marker)
	if mErr != nil {
		err = mErr
		return
	}

	_, vErr := writer.WriteString(card.loadVersion)
	if vErr != nil {
		err = vErr
		return
	}

	lErr := writer.WriteInt(card.language)
	if lErr != nil {
		err = lErr
		return
	}

	_, uErr := writer.WriteString(card.userID)
	if uErr != nil {
		err = uErr
		return
	}

	_, dErr := writer.WriteString(card.dataID)
	if dErr != nil {
		err = dErr
		return
	}

	keyArr := []string{"Custom", "Coordinate", "Parameter", "GameInfo", "Status", "Parameter2", "GameInfo2"}
	keyExtra := "KKEx"

	infoCount := len(card.infoHeader.lstInfo)
	lstInfo := make([]map[string]interface{}, infoCount)
	lstData := make(map[int][]byte)

	var i, d int
	var pos, datasz int64
	d = 0
	pos = int64(0)

	infoEx := card.findInfo(keyExtra)
	if infoEx.name == keyExtra {
		i++
	}

	for _, key := range keyArr {
		info := card.findInfo(key)
		if info.name == key {
			lstData[d] = card.data[key]
			size := int64(len(lstData[d]))

			infoMap := map[string]interface{}{
				"name":    fmt.Sprintf(info.name),
				"version": fmt.Sprintf("%s", info.version),
				"pos":     pos,
				"size":    size,
			}
			lstInfo[i] = infoMap

			pos += size
			datasz += size

			d++
			i++
		}
	}

	if infoEx.name == keyExtra {
		lstData[d] = card.data[keyExtra]
		size := int64(len(lstData[d]))

		infoMap := map[string]interface{}{
			"name":    fmt.Sprintf(infoEx.name),
			"version": fmt.Sprintf("%s", infoEx.version),
			"pos":     pos,
			"size":    size,
		}
		lstInfo[0] = infoMap

		d++
		pos += size
		datasz += size
	}

	blockHead := map[string][]map[string]interface{}{
		"lstInfo": lstInfo,
	}

	// j, jErr := json.Marshal(&blockHead)
	// if jErr == nil {
	// 	fmt.Println(string(j))
	// }

	head, msgErr := msgpack.Marshal(&blockHead)
	if msgErr != nil {
		err = msgErr
		return
	}

	headsz := len(head)
	headszErr := writer.WriteInt(int32(headsz))
	if headszErr != nil {
		err = headszErr
		return
	}

	_, headErr := writer.Write(head)
	if headErr != nil {
		err = headErr
		return
	}

	dataszErr := writer.WriteLong(datasz)
	if dataszErr != nil {
		err = dataszErr
		return
	}

	for j := 0; j < d; j++ {
		_, sErr := writer.Write(lstData[j])
		if sErr != nil {
			err = sErr
			return
		}
	}

	fluErr := writer.Flush()
	if fluErr != nil {
		err = fluErr
		return
	}

	re = true
	return
}

// WriteCharaFile implements for AISChara
func (sf *AISChara) WriteCharaFile(card AISCharaCard, filePath string) (re bool, err error) {
	re = false
	f, cErr := os.Create(filePath)
	if cErr != nil {
		err = cErr
		return
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	return sf.WriteChara(card, writer)
}

// ExtractChara implements for AISChara
func (sf *AISChara) ExtractChara(currDir string, flag int) (total int, write int, err error) {
	for k, v := range sf.card.charaCards {
		total++

		// Not male
		if flag == 1 && v.sex != 0 {
			continue
		}

		// Not female
		if flag == 2 && v.sex != 1 {
			continue
		}

		var c int
		var saveFilePath string

		saveFilePath = path.Join(currDir, k+".png")
		for {
			_, fErr := os.Stat(saveFilePath)
			if os.IsNotExist(fErr) {
				break
			}
			c++
			saveFilePath = path.Join(currDir, fmt.Sprintf("%s-%d.png", k, c))
		}

		_, saveErr := sf.WriteCharaFile(v, saveFilePath)
		if saveErr != nil {
			printError(saveErr)
		} else {
			write++
		}
	}
	return
}
