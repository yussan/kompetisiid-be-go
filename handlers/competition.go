package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	payloadModels "ki-be/models/payload"
	responsesModels "ki-be/models/response"
	tableModels "ki-be/models/tables"
	"ki-be/repositories"
	"ki-be/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func ListCompetition(c echo.Context) error {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var params = repositories.ParamsGetListCompetitions{
		Status:         c.QueryParam("status"),
		IsDraft:        c.QueryParam("is_draft"),
		IsGuaranted:    c.QueryParam("is_guaranted"),
		IsMediaPartner: c.QueryParam("is_mediapartner"),
		IsManage:       c.QueryParam("is_manage"),
		Username:       c.QueryParam("username"),
		Keyword:        c.QueryParam("keyword"),
		Tag:            c.QueryParam("tag"),
	}

	// get query page
	if c.QueryParam("page") != "" {
		pageNumber, _ := strconv.Atoi(c.QueryParam("page"))
		params.Page = pageNumber
	} else {
		params.Page = 1
	}

	// get query page
	if c.QueryParam("limit") != "" {
		limitNumber, _ := strconv.Atoi(c.QueryParam("limit"))
		params.Limit = limitNumber
	} else {
		params.Limit = 9
	}

	// get query by id main category
	if c.QueryParam("id_main_category") != "" {
		number, _ := strconv.Atoi(c.QueryParam("id_main_category"))
		params.IdMainCategory = number
	}

	// get query by id main category
	if c.QueryParam("main_category") != "" {
		params.MainCategory = c.QueryParam("main_category")
	}

	// get query by id sub category
	if c.QueryParam("id_sub_category") != "" {
		number, _ := strconv.Atoi(c.QueryParam("id_sub_category"))
		params.IdSubCategory = number
	}

	// get query by sub category
	if c.QueryParam("sub_category") != "" {
		params.SubCategory = c.QueryParam("sub_category")
	}

	// get query by status
	if c.QueryParam("status") != "" {
		params.Status = c.QueryParam("status")
	} else {
		params.Status = "posted"
	}

	data := repositories.GetCompetitions(c, params)
	total := repositories.GetCountCompetitions(c, params)
	status := 204
	message := "Kompetisi tidak ditemukan"

	if data != nil {
		status = 200
		message = "Success"
	}

	return c.JSON(http.StatusOK, responsesModels.GlobalResponse{Status: status, Message: message, Data: &echo.Map{"competitions": data, "total": total}})
}

func AddCompetition(c echo.Context) error {
	req := c.Request()

	// userKey validation
	userKey := req.Header.Get("userKey")

	if userKey == "" {
		return c.JSON(http.StatusBadRequest, responsesModels.GlobalResponse{Status: http.StatusForbidden, Message: "Please login first", Data: nil})
	} else {
		// check is available user with userKey
		_, userData := repositories.GetUserByUserKey(userKey)

		if userData.Id < 1 {
			return c.JSON(http.StatusBadRequest, responsesModels.GlobalResponse{Status: http.StatusForbidden, Message: "Please login first", Data: nil})
		} else {
			// -- add competition

			// receive body
			var payload payloadModels.PayloadCompetition
			err := json.NewDecoder(req.Body).Decode(&payload)

			if err != nil {
				fmt.Println(err)
				return c.JSON(http.StatusBadRequest, responsesModels.GlobalResponse{Status: http.StatusBadRequest, Message: "Error parsing payload", Data: nil})
			} else {
				now := time.Now()

				// upload to cloudinary first
				uploadDir := "/kompetisi-id/competition/" + userData.Username + "/" + fmt.Sprintf("%d", now.Year())

				_, uploadResult := utils.UploadCloudinary(uploadDir, payload.Poster)

				poster := uploadResult

				posterString, _ := json.Marshal(&poster)

				currentTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", now.Year(), now.Month(), now.Day(),
					now.Hour(), now.Minute(), now.Second())

				var isGuaranteed string = "0"
				if payload.IsGuaranteed == true {
					isGuaranteed = "1"
				}
				var isMediaPartner string = "0"
				if payload.IsMediaPartner == true {
					isMediaPartner = "1"
				}
				var isDraft string = "0"
				if payload.Draft == true {
					isDraft = "1"
				}
				new_data := tableModels.Kompetisi{
					Id_user:           userData.Id,
					Title:             payload.Title,
					Sort:              payload.Description,
					Poster:            "",
					Poster_cloudinary: string(posterString),
					Organizer:         payload.Organizer,
					DeadlineAt:        payload.DeadlineDate,
					AnnouncementAt:    payload.DeadlineDate,
					Id_main_cat:       payload.MainCat,
					Id_sub_cat:        payload.SubCat,
					Content:           payload.Content,
					PrizeTotal:        payload.PrizeTotal,
					PrizeDescription:  payload.PrizeDescription,
					Contact:           payload.Contacts,
					IsGuaranted:       isGuaranteed,
					IsMediaPartner:    isMediaPartner,
					IsManage:          "0",
					Draft:             isDraft,
					SourceLink:        payload.SourceLink,
					RegisterLink:      payload.RegisterLink,
					Announcements:     payload.Announcements,
					Tags:              payload.Tags,
					Status:            payload.Status,
					Views:             1,
					CreatedAt:         currentTime,
					UpdatedAt:         currentTime,
				}

				fmt.Println(new_data)

				errInsert, _ := repositories.WriteCompetition(c, new_data)

				if errInsert != nil {
					return c.JSON(http.StatusBadRequest, responsesModels.GlobalResponse{Status: http.StatusInternalServerError, Message: "Error insert ke DB", Data: nil})
				} else {
					return c.JSON(http.StatusBadRequest, responsesModels.GlobalResponse{Status: http.StatusOK, Message: "Sukses tambah kompetisi", Data: nil})
				}
			}
			// -- end of add competition
		}

	}

	return nil

}
