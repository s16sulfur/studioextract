package main

import (
	"errors"

	"github.com/sulfur/bbio"
)

func extractScene(currDir string, filePath string, flag int, full bool) (total int, write int, err error) {
	reader := bbio.NewReaderFile(filePath)
	var re, isHs, isKs bool

	pngSize := getPngSize(reader)

	if full {
		isPh := IsPHStudioSceneCard(reader)
		// PlayHome
		if isPh {
			phChara := NewPHChara()

			re, err = phChara.ReadScene(reader, pngSize)
			if !re {
				printError(errors.New("PHStudioSceneCard read failed"))
				return
			}

			total, write, err = phChara.ExtractChara(currDir, flag)
			return
		}

		isHs = IsHoneyStudioSceneCard(reader)
		isKs = IsKStudioSceneCard(reader)
	}

	isNeo := IsNeoSceneCard(reader)
	isNeoV2 := IsNeoV2SceneCard(reader)

	if isKs || isHs || isNeo || isNeoV2 {
		aisChara := NewAISChara()
		re, err = aisChara.ReadScene(reader, pngSize)
		if re {
			aisTotal, aisWrite, aisErr := aisChara.ExtractChara(currDir, flag)
			if aisErr != nil {
				err = aisErr
				return
			}

			total += aisTotal
			write += aisWrite
		}

		hsChara := NewHSChara()
		re, err = hsChara.ReadScene(reader, pngSize)
		if re {
			hsTotal, hsWrite, hsErr := hsChara.ExtractChara(currDir, flag)
			if hsErr != nil {
				err = hsErr
				return
			}

			total += hsTotal
			write += hsWrite
		}

		kkChara := NewKKChara()
		re, err = kkChara.ReadScene(reader, pngSize)
		if re {
			kkTotal, kkWrite, kkErr := kkChara.ExtractChara(currDir, flag)
			if kkErr != nil {
				err = kkErr
				return
			}

			total += kkTotal
			write += kkWrite
		}
	}

	return
}
