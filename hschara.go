package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sulfur/bbio"
)

func createSha256(data []byte, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return h.Sum(nil)
}

func readString(data []byte, offset int) (str string, n int, err error) {
	str = ""
	var count, shift, off, b int
	off = offset
	n = 0

	for {
		if shift == 5*7 {
			err = errors.New("Bad format 7Bit Int32")
			return
		}

		rb := data[off]
		n++
		off++

		b = int(rb)
		count |= (b & 0x7F) << shift
		shift += 7

		if (b & 0x80) == 0 {
			break
		}
	}

	strb := make([]byte, count)
	sn := copy(strb, data[off:])
	if sn == count {
		str = string(strb)
	}
	n += sn
	return
}

// HSHeaderInfo strcture
type HSHeaderInfo struct {
	name    string
	version int32
	pos     int64
	size    int64
}

// HSCharaCard strcture
type HSCharaCard struct {
	startOffset    int64
	sex            int32
	name           string
	pngSize        int64
	marker         string
	loadVersion    int32
	infoHeaderSize int32
	infoHeader     struct {
		lstInfo []HSHeaderInfo
	}
	dataSize int64
	data     map[string][]byte
}

func (sf *HSCharaCard) findInfo(name string) (info HSHeaderInfo) {
	for _, v := range sf.infoHeader.lstInfo {
		if v.name == name {
			info = v
			return
		}
	}
	return
}

func (sf *HSCharaCard) loadPreviewInfo() (err error) {
	sf.name = ""

	tagPreview := "プレビュー情報"
	info := sf.findInfo(tagPreview)
	if info.name == tagPreview {
		prevData := sf.data[info.name]

		var off int
		if info.version >= 4 {
			off += 4 // productNo
		}
		off += 4 // sex

		if info.version >= 2 {
			off += 8 // personality + nameLength

			str, _, err := readString(prevData, off)
			if err == nil {
				sf.name = str
			}
		}
	}
	return
}

// HSSceneCard strcture
type HSSceneCard struct {
	pngSize    int64
	charaCards map[string]HSCharaCard
}

// HSChara strcture
type HSChara struct {
	card *HSSceneCard
}

var honeyStudioMark = "【honey】"
var neoMark = "【-neo-】"
var hsCharaMaleMark = "【HoneySelectCharaMale】"
var hsCharaFemaleMark = "【HoneySelectCharaFemale】"

// NewHSChara implements for HSChara
func NewHSChara() *HSChara {
	c := &HSSceneCard{}
	return &HSChara{card: c}
}

// IsHoneyStudioSceneCard implements for Honey Studio scene
func IsHoneyStudioSceneCard(reader *bbio.Reader) bool {
	return (reader.Index([]byte(honeyStudioMark)) > 0)
}

// IsNeoSceneCard implements for neo scene
func IsNeoSceneCard(reader *bbio.Reader) bool {
	return (reader.Index([]byte(neoMark)) > 0)
}

// GenerateFileName implements for HSChara
func (sf *HSChara) GenerateFileName(sex int32) string {
	time.Sleep(2 * time.Millisecond)

	currentTime := time.Now()
	var fileName string
	if sex == 0 {
		fileName = "charaM_"
	} else if sex == 1 {
		fileName = "charaF_"
	} else {
		fileName = "chara_"
	}
	fileName += strings.ReplaceAll(currentTime.Format("2006.01.02.15.04.05.000"), ".", "")

	return fileName
}

// ReadChara implements for HSChara
func (sf *HSChara) ReadChara(reader *bbio.Reader, offset int64) (card HSCharaCard, err error) {
	startOffset, seekErr := reader.Seek(offset, io.SeekStart)
	if seekErr != nil {
		err = seekErr
		return
	}
	card.startOffset = int64(startOffset)

	mark, markErr := reader.ReadString()
	if markErr != nil {
		err = markErr
		return
	}
	if mark != hsCharaMaleMark && mark != hsCharaFemaleMark {
		err = errors.New("Honey Select Chara mark not found")
		return
	}
	card.marker = mark

	if mark == hsCharaMaleMark {
		card.sex = 0
	} else if mark == hsCharaFemaleMark {
		card.sex = 1
	}

	lvno, lvnoErr := reader.ReadInt32()
	if lvnoErr != nil {
		err = lvnoErr
		return
	}
	if lvno > 2 {
		err = errors.New("Version not supported")
		return
	}
	card.loadVersion = lvno

	headersz, headerszErr := reader.ReadInt32()
	if headerszErr != nil {
		err = headerszErr
		return
	}
	card.infoHeaderSize = headersz

	card.infoHeader.lstInfo = make([]HSHeaderInfo, headersz)
	for i := 0; i < int(headersz); i++ {
		tag, tagErr := reader.ReadStringFixed(128, true)
		if tagErr != nil {
			err = tagErr
			return
		}
		card.infoHeader.lstInfo[i].name = tag

		hver, hverErr := reader.ReadInt32()
		if hverErr != nil {
			err = hverErr
			return
		}
		card.infoHeader.lstInfo[i].version = hver

		hpos, hposErr := reader.ReadInt64()
		if hposErr != nil {
			err = hposErr
			return
		}
		card.infoHeader.lstInfo[i].pos = hpos

		hsz, hszErr := reader.ReadInt64()
		if hszErr != nil {
			err = hszErr
			return
		}
		card.infoHeader.lstInfo[i].size = hsz
	}

	dataOffset := reader.Position()
	card.data = make(map[string][]byte)
	infoCount := len(card.infoHeader.lstInfo)

	var rbsz int
	for i := 0; i < infoCount; i++ {
		info := card.infoHeader.lstInfo[i]
		sbOffset := dataOffset + info.pos

		_, sbErr := reader.Seek(sbOffset, io.SeekStart)
		if sbErr != nil {
			return card, sbErr
		}

		bBytes := make([]byte, info.size)
		rbn, rbErr := reader.Read(bBytes)
		if rbErr != nil {
			return card, rbErr
		}

		rbsz += rbn
		card.data[info.name] = bBytes
	}

	lOffset := dataOffset + int64(rbsz)
	_, lErr := reader.Seek(lOffset, io.SeekStart)
	if lErr == nil {
		var sigsz int
		sigsz = 16

		if lvno == 2 {
			sigsz += 32
		}

		sigBytes := make([]byte, sigsz)
		_, sigErr := reader.Read(sigBytes)
		if sigErr != nil {
			err = sigErr
		} else {
			card.data["sig"] = sigBytes
		}
	}

	loadErr := card.loadPreviewInfo()
	if loadErr != nil {
		err = loadErr
		return
	}

	return
}

// ReadScene implements for HSChara
func (sf *HSChara) ReadScene(reader *bbio.Reader, pngSize int64) (bool, error) {
	_, seekErr := reader.Seek(pngSize, io.SeekStart)
	if seekErr != nil {
		return false, seekErr
	}

	sf.card.pngSize = pngSize
	sf.card.charaCards = make(map[string]HSCharaCard)

	idxsMale := reader.FindAll([]byte(hsCharaMaleMark))
	for _, v := range idxsMale {
		chara, cerr := sf.ReadChara(reader, int64(v-1))
		if cerr != nil {
			if isDebug {
				printError(cerr)
			} else {
				printError(errors.New("Chara card read error"))
			}

		} else {
			charFileName := sf.GenerateFileName(chara.sex)
			sf.card.charaCards[charFileName] = chara
		}
	}

	idxsFemale := reader.FindAll([]byte(hsCharaFemaleMark))
	for _, v := range idxsFemale {
		chara, cerr := sf.ReadChara(reader, int64(v-1))
		if cerr != nil {
			if isDebug {
				printError(cerr)
			} else {
				printError(errors.New("Chara card read error"))
			}

		} else {
			charFileName := sf.GenerateFileName(chara.sex)
			sf.card.charaCards[charFileName] = chara
		}
	}

	return len(sf.card.charaCards) > 0, nil
}

// WriteChara implements for HSChara
func (sf *HSChara) WriteChara(card HSCharaCard, w io.Writer) (re bool, err error) {
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
	startPos := writer.Position()

	_, mErr := writer.WriteString(card.marker)
	if mErr != nil {
		err = mErr
		return
	}

	pnErr := writer.WriteInt(card.loadVersion)
	if pnErr != nil {
		err = pnErr
		return
	}

	infoCount := len(card.infoHeader.lstInfo)
	hsErr := writer.WriteInt(int32(infoCount))
	if hsErr != nil {
		err = hsErr
		return
	}

	lstData := make(map[int][]byte)
	var pos int64
	pos = int64(0)

	for i := 0; i < infoCount; i++ {
		info := card.infoHeader.lstInfo[i]
		tag := info.name
		lstData[i] = card.data[tag]
		size := int64(len(card.data[tag]))

		tagBytes := make([]byte, 128)
		copy(tagBytes, []byte(tag))

		_, tErr := writer.Write(tagBytes)
		if tErr != nil {
			err = tErr
			return
		}

		vErr := writer.WriteInt(info.version)
		if vErr != nil {
			err = vErr
			return
		}

		pErr := writer.WriteLong(pos)
		if pErr != nil {
			err = pErr
			return
		}

		sErr := writer.WriteLong(size)
		if sErr != nil {
			err = sErr
			return
		}

		pos += size
	}

	headerSize := writer.Position()
	for i := 0; i < infoCount; i++ {
		_, sErr := writer.Write(lstData[i])
		if sErr != nil {
			err = sErr
			return
		}
	}

	sha256Bytes := createSha256(pngBytes, card.marker)
	_, shaErr := writer.Write(sha256Bytes)
	if shaErr != nil {
		err = shaErr
		return
	}

	ppErr := writer.WriteLong(startPos)
	if ppErr != nil {
		err = ppErr
		return
	}

	hszErr := writer.WriteLong(headerSize)
	if hszErr != nil {
		err = hszErr
		return
	}

	fluErr := writer.Flush()
	if fluErr != nil {
		err = fluErr
		return
	}

	re = true
	return
}

// WriteCharaFile implements for HSChara
func (sf *HSChara) WriteCharaFile(card HSCharaCard, filePath string) (re bool, err error) {
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

// ExtractChara implements for HSChara
func (sf *HSChara) ExtractChara(currDir string, flag int) (total int, write int, err error) {
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
