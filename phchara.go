package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sulfur/bbio"
)

// PHCharaCard strcture
type PHCharaCard struct {
	sex       int32
	name      string
	version   int32
	sceneSex  int32
	hair      []byte
	head      []byte
	body      []byte
	wear      []byte
	accessory []byte
}

// PHSceneCard strcture
type PHSceneCard struct {
	pngSize    int64
	version    string
	charaCards map[string]PHCharaCard
}

var phStudioMark = "【PHStudio】"
var phCharaMaleMark = "【PlayHome_Male】"
var phCharaFemaleMark = "【PlayHome_Female】"

// PHChara strcture
type PHChara struct {
	card *PHSceneCard
}

func readPHColorHair(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()
	buf.PutInt(1) // colorType

	if version < 4 {
		bm, _ := reader.ReadBytes(16) // mainColor
		buf.Write(bm)

		// cuticleColor, cuticleExp
		buf.PutFloatAll(0.75, 0.75, 0.75, 1.0, 6.0)
		// fresnelColor, fresnelExp
		buf.PutFloatAll(0.75, 0.75, 0.75, 1.0, 0.3)

	} else {
		reader.ReadInt32() // colorType >> 1

		// mainColor, cuticleColor, cuticleExp, fresnelColor, fresnelExp
		bc, _ := reader.ReadBytes(56)
		buf.Write(bc)
	}

	b = buf.Bytes()
	return
}

func readPHColorPBR1(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()
	buf.PutInt(2) // colorType

	if version < 4 {
		bm, _ := reader.ReadBytes(16) // mainColor1
		buf.Write(bm)

		// specColor1
		buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
		// specular1, smooth1
		buf.PutFloatAll(0.0, 0.0)

	} else {
		colorType, _ := reader.ReadInt32() // colorType
		if colorType != 0 {
			// mainColor1, specColor1, specular1, smooth1
			bc, _ := reader.ReadBytes(40)
			buf.Write(bc)

		} else {
			// mainColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specular1, smooth1
			buf.PutFloatAll(0.0, 0.0)
		}
	}

	b = buf.Bytes()
	return
}

func readPHColorPBR2(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()
	buf.PutInt(3) // colorType

	if version < 4 {
		bm, _ := reader.ReadBytes(16) // mainColor1
		buf.Write(bm)

		// specColor1, specular1, smooth1
		buf.PutFloatAll(1.0, 1.0, 1.0, 1.0, 0.0, 0.0)
		// mainColor2
		buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
		// specColor2
		buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
		// specular1, smooth1
		buf.PutFloatAll(0.0, 0.0)

	} else {
		colorType, _ := reader.ReadInt32() // colorType
		if colorType != 0 {
			// mainColor1, specColor1, specular1
			// smooth1, mainColor2, specColor2
			bc, _ := reader.ReadBytes(72)
			buf.Write(bc)

			if version >= 5 {
				// specular2
				sp2, _ := reader.ReadSingle()
				buf.PutFloat(sp2)

			} else {
				// specular2
				buf.PutFloat(0.0)
			}

			// smooth2
			sm2, _ := reader.ReadSingle()
			buf.PutFloat(sm2)

		} else {
			// mainColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specular1, smooth1
			buf.PutFloatAll(0.0, 0.0)

			// mainColor2
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specColor2
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specular2, smooth2
			buf.PutFloatAll(0.0, 0.0)
		}
	}

	b = buf.Bytes()
	return
}

func readPHColorAlloy(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()
	buf.PutInt(4) // colorType

	if version < 4 {
		bm, _ := reader.ReadBytes(16) // mainColor
		buf.Write(bm)

		// metallic, smooth
		buf.PutFloatAll(0.0, 0.0)

	} else {
		colorType, _ := reader.ReadInt32() // colorType
		if colorType != 0 {
			// mainColor, metallic, smooth
			bc, _ := reader.ReadBytes(24)
			buf.Write(bc)

		} else {
			// mainColor, metallic, smooth
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0, 0.0, 0.0)
		}
	}

	b = buf.Bytes()
	return
}

func readPHColorAlloyHSVOffset(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()
	buf.PutInt(5) // colorType

	if version < 4 {
		reader.ReadBytes(16)

		// hsv offset + alpha, metallic, smooth
		buf.PutFloatAll(0.0, 1.0, 1.0, 1.0, 0.0, 0.562)

	} else {
		colorType, _ := reader.ReadInt32() // colorType
		if colorType != 0 {

			if version < 6 {
				reader.ReadBytes(16)
				// hsv offset + alpha
				buf.PutFloatAll(0.0, 1.0, 1.0, 1.0)

			} else {
				// offset_h, offset_s, offset_v
				bm, _ := reader.ReadBytes(12)
				buf.Write(bm)

				if version == 7 {
					reader.ReadBoolean()
					// alpha
					a, _ := reader.ReadSingle()
					buf.PutFloat(a)

				} else if version >= 8 {
					// alpha
					a, _ := reader.ReadSingle()
					buf.PutFloat(a)

				} else {
					// alpha
					buf.PutFloat(1.0)
				}
			}

			// metallic, smooth
			ms, _ := reader.ReadBytes(8)
			buf.Write(ms)

		} else {
			// hsv offset + alpha, metallic, smooth
			buf.PutFloatAll(0.0, 1.0, 1.0, 1.0, 0.0, 0.562)
		}
	}

	b = buf.Bytes()
	return
}

func readPHColorEyeHighlight(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()
	buf.PutInt(7) // colorType

	if version < 4 {
		bm, _ := reader.ReadBytes(16) // mainColor1
		buf.Write(bm)

		// specColor1
		buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
		// specular1, smooth1
		buf.PutFloatAll(0.0, 0.0)

	} else {
		colorType, _ := reader.ReadInt32() // colorType
		if colorType != 0 {
			// mainColor1, specColor1, specular1, smooth1
			bc, _ := reader.ReadBytes(40)
			buf.Write(bc)

		} else {
			// mainColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0)
			// specular1, smooth1
			buf.PutFloatAll(0.0, 0.0)
		}
	}

	b = buf.Bytes()
	return
}

func readPHHair(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()

	partsCount, _ := reader.ReadInt32()
	buf.PutInt(partsCount)

	for i := 0; i < int(partsCount); i++ {
		// hairPartID
		id, _ := reader.ReadInt32()
		buf.PutInt(id)

		// hairColor
		hairColor, _ := readPHColorHair(reader, version)
		buf.Write(hairColor)

		if version > 0 {
			// acceColor
			acceColor, _ := readPHColorPBR1(reader, version)
			buf.Write(acceColor)

		} else {
			// acceColor
			buf.Write([]byte{0x0, 0x0, 0x0, 0x0})
		}
	}

	b = buf.Bytes()
	return
}

func readPHHead(reader *bbio.Reader, version int32, sex int32) (b []byte, err error) {
	buf := bbio.NewBuffer()

	// headID, faceTexID, detailID, detailWeight, eyeBrowID
	b1, _ := reader.ReadBytes(20)
	buf.Write(b1)

	// eyeBrowColor
	ebc, _ := readPHColorPBR1(reader, version)
	buf.Write(ebc)

	if version < 4 {
		// eyeScleraColor
		esc, _ := reader.ReadBytes(16)

		// eyeID_L
		eyeIDL, _ := reader.ReadInt32()
		buf.PutInt(eyeIDL)

		// eyeScleraColorL
		buf.Write(esc)

		// eyeIrisColorL
		eicl, _ := reader.ReadBytes(16)
		buf.Write(eicl)

		// eyePupilDilationL, eyeEmissiveL
		buf.PutFloatAll(0.0, 0.5)

		// eyeID_R
		eyeIDR, _ := reader.ReadInt32()
		buf.PutInt(eyeIDR)

		// eyeScleraColorR
		buf.Write(esc)

		// eyeIrisColorR
		eicr, _ := reader.ReadBytes(16)
		buf.Write(eicr)

		// eyePupilDilationR, eyeEmissiveR
		buf.PutFloatAll(0.0, 0.5)

	} else {
		// eyeID_L, eyeScleraColorL, eyeIrisColorL eyePupilDilationL
		bel, _ := reader.ReadBytes(40)
		buf.Write(bel)

		if version >= 10 {
			// eyeEmissiveL
			eel, _ := reader.ReadSingle()
			buf.PutFloat(eel)

		} else {
			// eyeEmissiveL
			buf.PutFloat(0.5)
		}

		// eyeID_R, eyeScleraColorR, eyeIrisColorR, eyePupilDilationR
		ber, _ := reader.ReadBytes(40)
		buf.Write(ber)

		if version >= 10 {
			// eyeEmissiveR
			eer, _ := reader.ReadSingle()
			buf.PutFloat(eer)

		} else {
			// eyeEmissiveR
			buf.PutFloat(0.5)
		}
	}

	// tattooID, tattooColor
	bt, _ := reader.ReadBytes(20)
	buf.Write(bt)

	// shapeVals
	shapeCount, _ := reader.ReadInt32()
	buf.PutInt(shapeCount)
	sh, _ := reader.ReadBytes(int(shapeCount) * 4)
	buf.Write(sh)

	if sex == 0 {
		// eyeLash
		el, _ := reader.ReadInt32()
		buf.PutInt(el)
		elb, _ := readPHColorPBR1(reader, version)
		buf.Write(elb)

		// eyeshadow
		eyeshadow, _ := reader.ReadBytes(20)
		buf.Write(eyeshadow)

		// cheek
		cheek, _ := reader.ReadBytes(20)
		buf.Write(cheek)

		// lip
		lip, _ := reader.ReadBytes(20)
		buf.Write(lip)

		// mole
		mole, _ := reader.ReadBytes(20)
		buf.Write(mole)

		// eyeHighlight
		eyehl, _ := reader.ReadInt32()
		buf.PutInt(eyehl)
		eyehlb, _ := readPHColorEyeHighlight(reader, version)
		buf.Write(eyehlb)

	} else {
		// beard
		beard, _ := reader.ReadBytes(20)
		buf.Write(beard)

		if version >= 2 {
			// eyeHighlightColor
			ehlc, _ := readPHColorEyeHighlight(reader, version)
			buf.Write(ehlc)

		} else {
			// eyeHighlightColor
			buf.PutInt(7)                       // colorType
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0) // mainColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0) // specColor1
			buf.PutFloatAll(0.0, 0.0)           // specular1, smooth1
		}
	}

	b = buf.Bytes()
	return
}

func readPHBody(reader *bbio.Reader, version int32, sex int32) (b []byte, err error) {
	buf := bbio.NewBuffer()

	// bodyID
	bID, _ := reader.ReadInt32()
	buf.PutInt(bID)

	// skinColor
	sc, _ := readPHColorAlloyHSVOffset(reader, version)
	buf.Write(sc)

	// detailID, detailWeight, underhairID
	b1, _ := reader.ReadBytes(12)
	buf.Write(b1)

	// underhairColor
	uhc, _ := readPHColorAlloy(reader, version)
	buf.Write(uhc)

	// tattooID, tattooColor
	bt, _ := reader.ReadBytes(20)
	buf.Write(bt)

	// shapeVals
	shapeCount, _ := reader.ReadInt32()
	buf.PutInt(shapeCount)
	sh, _ := reader.ReadBytes(int(shapeCount) * 4)
	buf.Write(sh)

	if sex == 0 {
		// nipID
		nipID, _ := reader.ReadInt32()
		buf.PutInt(nipID)

		// nipColor
		nipColor, _ := readPHColorAlloyHSVOffset(reader, version)
		buf.Write(nipColor)

		// sunburnID, sunburnColor
		sun, _ := reader.ReadBytes(20)
		buf.Write(sun)

		if version >= 3 {
			// nailColor
			nail, _ := readPHColorAlloyHSVOffset(reader, version)
			buf.Write(nail)

			if version >= 9 {
				// manicureColor
				manc, _ := readPHColorPBR1(reader, version)
				buf.Write(manc)

			} else {
				// manicureColor
				buf.PutInt(2)                       // colorType
				buf.PutFloatAll(1.0, 1.0, 1.0, 0.0) // mainColor1
				buf.PutFloatAll(1.0, 1.0, 1.0, 1.0) // specColor1
				buf.PutFloatAll(0.0, 0.0)           // specular1, smooth1
			}

			// areolaSize, bustSoftness, bustWeight
			abb, _ := reader.ReadBytes(12)
			buf.Write(abb)

		} else {
			// nailColor
			buf.PutInt(5)                       // colorType
			buf.PutFloatAll(0.0, 1.0, 1.0, 1.0) // hsv offset + alpha
			buf.PutFloatAll(0.0, 0.562)         // metallic, smooth

			// manicureColor
			buf.PutInt(2)                       // colorType
			buf.PutFloatAll(1.0, 1.0, 1.0, 0.0) // mainColor1
			buf.PutFloatAll(1.0, 1.0, 1.0, 1.0) // specColor1
			buf.PutFloatAll(0.0, 0.0)           // specular1, smooth1

			// areolaSize, bustSoftness, bustWeight
			buf.PutFloatAll(0.5, 0.5, 0.5)
		}
	}

	b = buf.Bytes()
	return
}

func readPHWear(reader *bbio.Reader, version int32, sex int32) (b []byte, err error) {
	buf := bbio.NewBuffer()

	for i := 0; i < 11; i++ {
		// WEAR_TYPE, id
		bb, _ := reader.ReadBytes(8)
		buf.Write(bb)

		// color
		bc, _ := readPHColorPBR2(reader, version)
		buf.Write(bc)
	}

	if sex == 0 {
		// isSwimwear, swimOptTop, swimOptBtm
		bf, _ := reader.ReadBytes(3)
		buf.Write(bf)
	}

	b = buf.Bytes()
	return
}

func readPHAccessory(reader *bbio.Reader, version int32) (b []byte, err error) {
	buf := bbio.NewBuffer()

	for i := 0; i < 10; i++ {
		// ACCESSORY_TYPE, id, nowAttach, addPos, addRot addScl
		bb, _ := reader.ReadBytes(48)
		buf.Write(bb)

		// color
		bc, _ := readPHColorPBR2(reader, version)
		buf.Write(bc)
	}

	b = buf.Bytes()
	return
}

func readPHCustomParameter(reader *bbio.Reader) (card PHCharaCard, err error) {
	ver, vErr := reader.ReadInt32()
	if vErr != nil {
		err = vErr
		return
	}
	card.version = ver

	sex, sexErr := reader.ReadInt32()
	if sexErr != nil {
		err = sexErr
		return
	}
	card.sex = sex

	hair, haErr := readPHHair(reader, ver)
	if haErr != nil {
		err = haErr
		return
	}
	card.hair = hair

	head, heErr := readPHHead(reader, ver, sex)
	if heErr != nil {
		err = heErr
		return
	}
	card.head = head

	body, bErr := readPHBody(reader, ver, sex)
	if bErr != nil {
		err = bErr
		return
	}
	card.body = body

	wear, wErr := readPHWear(reader, ver, sex)
	if wErr != nil {
		err = wErr
		return
	}
	card.wear = wear

	accessory, aErr := readPHAccessory(reader, ver)
	if aErr != nil {
		err = aErr
		return
	}
	card.accessory = accessory
	return
}

func readPHChild(reader *bbio.Reader, version int, lstChara map[int]PHCharaCard) (err error) {
	chCount, ccErr := reader.ReadInt32()
	if ccErr != nil {
		err = ccErr
		return
	}
	var readErr error
	for i := 0; i < int(chCount); i++ {
		iType, itErr := reader.ReadInt32()
		if itErr != nil {
			err = itErr
			return
		}
		switch iType {
		case 0:
			readErr = readPHOICharInfo(reader, version, lstChara)
		case 1:
			readErr = readPHOIItemInfo(reader, version, lstChara)
		case 2:
			readErr = readPHOILightInfo(reader)
		case 3:
			readErr = readPHOIFolderInfo(reader, version, lstChara)
		default:
		}

		if readErr != nil {
			err = readErr
			return
		}
	}
	return
}

func readPHCharFileStatus(reader *bbio.Reader, version int) (name string, err error) {
	reader.ReadBytes(4) // coordinateType

	countAccessory, _ := reader.ReadInt32()
	reader.ReadBytes(int(countAccessory))

	// eyesPtn, eyesOpen, eyesOpenMin, eyesOpenMax, eyesFixed
	// mouthPtn, mouthOpen, mouthOpenMin, mouthOpenMax, mouthFixed
	// tongueState, eyesLookPtn, eyesTargetNo, eyesTargetRate
	// neckLookPtn, neckTargetNo, neckTargetRate, eyesBlink, disableShapeMouth
	reader.ReadBytes(67)

	countClothesState, _ := reader.ReadInt32()
	reader.ReadBytes(int(countClothesState))

	countSiruLv, _ := reader.ReadInt32()
	reader.ReadBytes(int(countSiruLv))

	// nipStand, hohoAkaRate, tearsLv
	// disableShapeBustL, disableShapeBustR, disableShapeNipL
	// disableShapeNipR, hideEyesHighlight
	reader.ReadBytes(17)

	name = ""
	if version >= 14 {
		name, _ = reader.ReadString()
	}
	return
}

func readPHObjectInfo(reader *bbio.Reader, other bool) (err error) {
	reader.ReadBytes(4) // dicKey
	reader.ReadString() // ChangeAmount.pos
	reader.ReadString() // ChangeAmount.rot
	reader.ReadString() // ChangeAmount.scale
	if other {
		// treeState, visible
		reader.ReadBytes(5)
	}
	return
}

func readPHOICharInfo(reader *bbio.Reader, version int, lstChara map[int]PHCharaCard) (err error) {
	err = readPHObjectInfo(reader, true)
	if err != nil {
		return
	}

	sex, sexErr := reader.ReadInt32()
	if sexErr != nil {
		err = sexErr
		return
	}

	// CustomParameter
	card, cErr := readPHCustomParameter(reader)
	if cErr != nil {
		err = cErr
		return
	}
	card.sceneSex = sex

	// CharFileStatus
	name, cfsErr := readPHCharFileStatus(reader, version)
	if cfsErr != nil {
		err = cfsErr
		return
	}
	card.name = name

	idx := len(lstChara)
	lstChara[idx] = card

	// bones
	countBones, _ := reader.ReadInt32()
	for i := 0; i < int(countBones); i++ {
		reader.ReadBytes(4)
		readPHObjectInfo(reader, false)
	}

	// IkTarget
	countIkTarget, _ := reader.ReadInt32()
	for i := 0; i < int(countIkTarget); i++ {
		reader.ReadBytes(4)
		readPHObjectInfo(reader, false)
	}

	// child
	countChild, _ := reader.ReadInt32()
	for i := 0; i < int(countChild); i++ {
		reader.ReadBytes(4)
		readPHChild(reader, version, lstChara)
	}

	// kinematicMode, animeInfo.group, animeInfo.category, animeInfo.no
	// handPtnL, handPtnR, skinRate nipple, siru
	reader.ReadBytes(37)

	if version >= 12 {
		// faceOption
		reader.ReadBytes(4)
	}

	// mouthOpen, lipSync
	reader.ReadBytes(5)

	// lookAtTarget
	readPHObjectInfo(reader, false)

	// enableIK, activeIK, enableFK, activeFK
	// expression, animeSpeed
	reader.ReadBytes(22)

	if version < 12 {
		// animePattern[0]
		reader.ReadSingle()
	} else {
		// animePattern[0], animePattern[1]
		reader.ReadBytes(8)
	}

	// animeOptionVisible, isAnimeForceLoop
	reader.ReadBytes(2)

	// voice
	voiceCount, _ := reader.ReadInt32()
	for i := 0; i < int(voiceCount); i++ {
		// group, category, no
		reader.ReadBytes(12)
	}
	// repeat
	reader.ReadBytes(4)

	if sex == 0 {
		// visibleSimple
		reader.ReadBoolean()
		// simpleColor
		reader.ReadString()
		// visibleSon, animeOptionParam[0], animeOptionParam[1]
		reader.ReadBytes(9)
	}

	// nectState
	countNeckByte, _ := reader.ReadInt32()
	reader.ReadBytes(int(countNeckByte))

	// eyesState
	countEyesByte, _ := reader.ReadInt32()
	reader.ReadBytes(int(countEyesByte))

	// animeNormalizedTime
	reader.ReadBytes(4)

	// AccessGroup state
	countAccessGroup, _ := reader.ReadInt32()
	for i := 0; i < int(countAccessGroup); i++ {
		// key, TreeState
		reader.ReadBytes(8)
	}

	// AccessNo state
	countAccessNo, _ := reader.ReadInt32()
	for i := 0; i < int(countAccessNo); i++ {
		// key, TreeState
		reader.ReadBytes(8)
	}

	return
}

func readPHOIItemInfo(reader *bbio.Reader, version int, lstChara map[int]PHCharaCard) (err error) {
	err = readPHObjectInfo(reader, true)
	if err != nil {
		return
	}

	// no, animeSpeed, colortype, color, color2, enableFK
	reader.ReadBytes(157)

	// bones
	cBone, err := reader.ReadInt32()
	for i := 0; i < int(cBone); i++ {
		reader.ReadString()
		err = readPHObjectInfo(reader, false)
	}

	// animeNormalizedTime
	reader.ReadBytes(4)
	err = readPHChild(reader, version, lstChara)
	return
}

func readPHOILightInfo(reader *bbio.Reader) (err error) {
	err = readPHObjectInfo(reader, true)
	if err != nil {
		return
	}
	// no, color, intensity, range, spotAngle, shadow, enable, drawTarget
	reader.ReadBytes(35)
	return
}

func readPHOIFolderInfo(reader *bbio.Reader, version int, lstChara map[int]PHCharaCard) (err error) {
	err = readPHObjectInfo(reader, true)
	if err != nil {
		return
	}
	// name
	reader.ReadString()
	err = readPHChild(reader, version, lstChara)
	return
}

// NewPHChara implements for PHChara
func NewPHChara() *PHChara {
	c := &PHSceneCard{}
	return &PHChara{card: c}
}

// IsPHStudioSceneCard implements for PH studio scene
func IsPHStudioSceneCard(reader *bbio.Reader) bool {
	return (reader.Index([]byte(phStudioMark)) > 0)
}

// GenerateFileName implements for PHChara
func (sf *PHChara) GenerateFileName(name string) string {
	return name + ".png"
}

// ReadScene implements for PHChara
func (sf *PHChara) ReadScene(reader *bbio.Reader, pngSize int64) (re bool, err error) {
	_, seekErr := reader.Seek(pngSize, io.SeekStart)
	if seekErr != nil {
		err = seekErr
		return
	}

	sf.card.pngSize = pngSize
	sf.card.charaCards = make(map[string]PHCharaCard)

	ver, vErr := reader.ReadString()
	if vErr != nil {
		err = vErr
		return
	}
	sf.card.version = ver

	vTmp := strings.ReplaceAll(ver, ".", "")
	iVer, ivErr := strconv.ParseInt(vTmp, 16, 64)
	if ivErr != nil {
		err = ivErr
		return
	}
	verInt := int(iVer)

	iCount, icErr := reader.ReadInt32()
	if icErr != nil {
		err = icErr
		return
	}

	lstChara := make(map[int]PHCharaCard)

	var readErr error
	for i := 0; i < int(iCount); i++ {
		reader.ReadInt32()

		iType, itErr := reader.ReadInt32()
		if itErr != nil {
			err = itErr
			return
		}
		switch iType {
		case 0:
			readErr = readPHOICharInfo(reader, verInt, lstChara)
		case 1:
			readErr = readPHOIItemInfo(reader, verInt, lstChara)
		case 2:
			readErr = readPHOILightInfo(reader)
		case 3:
			readErr = readPHOIFolderInfo(reader, verInt, lstChara)
		default:
		}

		if readErr != nil {
			err = readErr
			return
		}
	}

	cCount := len(lstChara)
	if cCount > 0 {
		for i := 0; i < cCount; i++ {
			chara := lstChara[i]
			var key, name string
			var c int

			if chara.name != "" {
				name = chara.name
			} else {
				name = strings.ReplaceAll(time.Now().Format("2006.01.02.15.04.05.000"), ".", "")
			}
			key = name

			for {
				if _, ok := sf.card.charaCards[key]; ok {
					key = fmt.Sprintf("%s (%d)", name, c)
					c++
				} else {
					break
				}
			}
			sf.card.charaCards[key] = chara
		}
	}

	re = true
	return
}

// WriteChara implements for PHChara
func (sf *PHChara) WriteChara(card PHCharaCard, w io.Writer) (re bool, err error) {
	re = false
	writer := bbio.NewWriter(w)

	var sex int
	if card.sex == 0 {
		sex = 1
	} else {
		sex = 0
	}

	pngBytes, pngErr := createPng(252, 352, sex)
	if pngErr != nil {
		err = pngErr
		return
	}

	_, pwErr := writer.Write(pngBytes)
	if pwErr != nil {
		err = pwErr
		return
	}

	if card.sex == 0 {
		writer.WriteString(phCharaFemaleMark)
	} else {
		writer.WriteString(phCharaMaleMark)
	}

	writer.WriteInt(10)
	writer.WriteInt(card.sex)
	writer.Write(card.hair)
	writer.Write(card.head)
	writer.Write(card.body)
	writer.Write(card.wear)
	writer.Write(card.accessory)

	fluErr := writer.Flush()
	if fluErr != nil {
		err = fluErr
		return
	}

	re = true
	return
}

// WriteCharaFile implements for PHChara
func (sf *PHChara) WriteCharaFile(card PHCharaCard, filePath string) (re bool, err error) {
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

// ExtractChara implements for PHChara
func (sf *PHChara) ExtractChara(currDir string, flag int) (total int, write int, err error) {
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
