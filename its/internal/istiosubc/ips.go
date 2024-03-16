// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	//"fmt"
	"strings"
	"strconv"
	"net"
	pkgerrors "github.com/pkg/errors"
)

type newIP struct {
	Ip net.IP
}

func (n* newIP) getIpAddress()(net.IP, error) {

	ip := n.Ip
	sa := strings.Split(ip.String(), ".")
	i, err := strconv.Atoi(sa[3])
	if err != nil {
		return net.IP{}, err
	}
	if  i >= 255  {
		return net.IP{}, pkgerrors.New("Out of ip address")
	}

	nip := strconv.Itoa(i+1)
	sa[3] = nip
	newip := sa[0] + "." + sa[1] + "." + sa[2] + "." + sa[3]
	n.Ip = net.ParseIP(newip)

	return n.Ip, nil
}
