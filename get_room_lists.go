package ews

import "encoding/xml"

type GetRoomListsRequest struct {
	XMLName struct{} `xml:"m:GetRoomLists"`
}

type GetRoomListsResponse struct {
	ResponseClass string    `xml:"ResponseClass,attr"`
	ResponseCode  string    `xml:"ResponseCode"`
	RoomLists     RoomLists `xml:"RoomLists"`
}

type RoomLists struct {
	Address []Address `xml:"Address"`
}

type Address struct {
	Name         string `xml:"Name"`
	EmailAddress string `xml:"EmailAddress"`
	RoutingType  string `xml:"RoutingType"`
	MailboxType  string `xml:"MailboxType"`
	ItemId       ItemId `xml:"ItemId"`
}

type ItemId struct {
	Id        string `xml:"Id,attr"`
	ChangeKey string `xml:"ChangeKey,attr"`
}

type getRoomListsResponseEnvelop struct {
	XMLName struct{}                 `xml:"Envelope"`
	Body    getRoomListsResponseBody `xml:"Body"`
}
type getRoomListsResponseBody struct {
	GetRoomListsResponse GetRoomListsResponse `xml:"GetRoomListsResponse"`
}

func GetRoomLists(c *Client) (*GetRoomListsResponse, error) {

	xmlBytes, err := xml.MarshalIndent(&GetRoomListsRequest{}, "", "  ")
	if err != nil {
		return nil, err
	}

	bb, err := c.sendAndReceive(xmlBytes)
	if err != nil {
		return nil, err
	}

	var soapResp getRoomListsResponseEnvelop
	err = xml.Unmarshal(bb, &soapResp)
	if err != nil {
		return nil, err
	}

	return &soapResp.Body.GetRoomListsResponse, nil
}