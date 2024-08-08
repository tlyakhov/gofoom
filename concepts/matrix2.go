// Copyright (c) Tim Lyakhovetskiy
// SPDX-License-Identifier: MPL-2.0

package concepts

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Matrix2 [6]float64

var IdentityMatrix2 = Matrix2{1, 0, 0, 1, 0, 0}

func (m *Matrix2) SetIdentity() {
	m[3] = 1
	m[2] = 0
	m[1] = 0
	m[0] = 1
}

func (m *Matrix2) From(m2 *Matrix2) {
	m[5] = m2[5]
	m[4] = m2[4]
	m[3] = m2[3]
	m[2] = m2[2]
	m[1] = m2[1]
	m[0] = m2[0]
}

func (m *Matrix2) IsIdentity() bool {
	return m[0] == 1 && m[2] == 0 && m[4] == 0 &&
		m[1] == 0 && m[3] == 1 && m[5] == 0
}

func (m *Matrix2) String() string {
	return fmt.Sprintf(
		"%v, %v, %v, %v, %v, %v",
		m[0], m[2], m[4],
		m[1], m[3], m[5],
	)
}

func (m *Matrix2) StringHuman() string {
	return fmt.Sprintf(
		"[[%.2f %.2f %.2f] [%.2f %.2f %.2f]]",
		m[0], m[2], m[4],
		m[1], m[3], m[5],
	)
}

func (m *Matrix2) TranslateSelf(delta *Vector2) *Matrix2 {
	m[4] += delta[0]
	m[5] += delta[1]
	return m
}

func (m *Matrix2) AxisScale(scale *Vector2) *Matrix2 {
	m[0], m[2], m[4] = m[0]*scale[0], m[2]*scale[0], m[4]*scale[0]
	m[1], m[3], m[5] = m[1]*scale[1], m[3]*scale[1], m[5]*scale[1]
	return m
}

func (m *Matrix2) AxisScaleBasis(basis *Vector2, scale *Vector2) *Matrix2 {
	m[4], m[5] = m[4]-basis[0], m[5]-basis[1]
	m[0], m[2], m[4] = m[0]*scale[0], m[2]*scale[0], m[4]*scale[0]
	m[1], m[3], m[5] = m[1]*scale[1], m[3]*scale[1], m[5]*scale[1]
	m[4], m[5] = m[4]+basis[0], m[5]+basis[1]
	return m
}

func (m *Matrix2) ScaleSelf(scale float64) *Matrix2 {
	m[0], m[2], m[4] = m[0]*scale, m[2]*scale, m[4]*scale
	m[1], m[3], m[5] = m[1]*scale, m[3]*scale, m[5]*scale
	return m
}

func (m *Matrix2) ScaleBasis(basis *Vector2, scale float64) *Matrix2 {
	m[4], m[5] = m[4]-basis[0], m[5]-basis[1]
	m[0], m[2], m[4] = m[0]*scale, m[2]*scale, m[4]*scale
	m[1], m[3], m[5] = m[1]*scale, m[3]*scale, m[5]*scale
	m[4], m[5] = m[4]+basis[0], m[5]+basis[1]
	return m
}

func (m *Matrix2) Rotate(angle float64) *Matrix2 {
	sint, cost := math.Sincos(angle * Deg2rad)
	m = m.Mul(&Matrix2{cost, sint, -sint, cost, 0, 0})
	return m
}

func (m *Matrix2) RotateSelf(angle float64) *Matrix2 {
	sint, cost := math.Sincos(angle * Deg2rad)
	m.MulSelf(&Matrix2{cost, sint, -sint, cost, 0, 0})
	return m
}

func (m *Matrix2) RotateBasis(basis *Vector2, angle float64) *Matrix2 {
	sint, cost := math.Sincos(angle)
	m[4], m[5] = m[4]-basis[0], m[5]-basis[1]
	m = m.Mul(&Matrix2{cost, sint, -sint, cost, 0, 0})
	m[4], m[5] = m[4]+basis[0], m[5]+basis[1]
	return m
}

func (m *Matrix2) Mul(next *Matrix2) *Matrix2 {
	return &Matrix2{
		next[0]*m[0] + next[2]*m[1],
		next[1]*m[0] + next[3]*m[1],
		next[0]*m[2] + next[2]*m[3],
		next[1]*m[2] + next[3]*m[3],
		next[0]*m[4] + next[2]*m[5] + next[4],
		next[1]*m[4] + next[3]*m[5] + next[5],
	}
}

func (m *Matrix2) MulSelf(next *Matrix2) *Matrix2 {
	m[0], m[1] = next[0]*m[0]+next[2]*m[1], next[1]*m[0]+next[3]*m[1]
	m[2], m[3] = next[0]*m[2]+next[2]*m[3], next[1]*m[2]+next[3]*m[3]
	m[4], m[5] = next[0]*m[4]+next[2]*m[5]+next[4], next[1]*m[4]+next[3]*m[5]+next[5]
	return m
}

func (m *Matrix2) Project(u *Vector2) *Vector2 {
	return &Vector2{
		m[0]*u[0] + m[2]*u[1] + m[4],
		m[1]*u[0] + m[3]*u[1] + m[5]}
}

func (m *Matrix2) ProjectXY(u *Vector3) *Vector3 {
	return &Vector3{
		m[0]*u[0] + m[2]*u[1] + m[4],
		m[1]*u[0] + m[3]*u[1] + m[5],
		u[2],
	}
}

func (m *Matrix2) ProjectXZ(u *Vector3) *Vector3 {
	return &Vector3{
		m[0]*u[0] + m[2]*u[2] + m[4],
		u[1],
		m[1]*u[0] + m[3]*u[2] + m[5]}
}

func (m *Matrix2) ProjectSelf(u *Vector2) *Vector2 {
	u[0], u[1] = m[0]*u[0]+m[2]*u[1]+m[4], m[1]*u[0]+m[3]*u[1]+m[5]
	return u
}

func (m Matrix2) ProjectXYSelf(u *Vector3) *Vector3 {
	u[0], u[1] = m[0]*u[0]+m[2]*u[1]+m[4], m[1]*u[0]+m[3]*u[1]+m[5]
	return u
}

func (m *Matrix2) ProjectXZSelf(u *Vector3) *Vector3 {
	u[0], u[2] = m[0]*u[0]+m[2]*u[2]+m[4], m[1]*u[0]+m[3]*u[2]+m[5]
	return u
}

func (m *Matrix2) Unproject(u *Vector2) *Vector2 {
	det := m[0]*m[3] - m[2]*m[1]
	return &Vector2{
		(m[3]*(u[0]-m[4]) - m[2]*(u[1]-m[5])) / det,
		(-m[1]*(u[0]-m[4]) + m[0]*(u[1]-m[5])) / det,
	}
}

func (m Matrix2) GetTransform() (angle float64, translation Vector2, scale Vector2) {
	basis1 := &Vector2{m[0], m[1]}
	basis2 := &Vector2{m[2], m[3]}
	scale[0] = basis1.Length()
	scale[1] = basis2.Length()
	translation[0] = m[4]
	translation[1] = m[5]
	angle = math.Atan2(basis1[1], basis1[0]) + math.Atan2(basis2[1], basis2[0]) - math.Pi*0.5
	return
}

func (m *Matrix2) Deserialize(data []any) {
	for i, v := range data {
		if i >= 6 {
			break
		}
		m[i] = v.(float64)
	}
}

func (m *Matrix2) Serialize() [6]float64 {
	return ([6]float64)(*m)
}

// ParseMatrix2 parses strings in the form "x,y,z,a,b,c" into matrices.
func ParseMatrix2(s string) (*Matrix2, error) {
	result := &Matrix2{}
	split := strings.Split(s, ",")
	if len(split) != 6 {
		return result, errors.New("can't parse Matrix2: input string should have six comma-separated values")
	}
	var err error
	result[0], err = strconv.ParseFloat(strings.TrimSpace(split[0]), 64)
	if err != nil {
		return result, err
	}
	result[2], err = strconv.ParseFloat(strings.TrimSpace(split[1]), 64)
	if err != nil {
		return result, err
	}
	result[4], err = strconv.ParseFloat(strings.TrimSpace(split[2]), 64)
	if err != nil {
		return result, err
	}
	result[1], err = strconv.ParseFloat(strings.TrimSpace(split[3]), 64)
	if err != nil {
		return result, err
	}
	result[3], err = strconv.ParseFloat(strings.TrimSpace(split[4]), 64)
	if err != nil {
		return result, err
	}
	result[5], err = strconv.ParseFloat(strings.TrimSpace(split[5]), 64)
	if err != nil {
		return result, err
	}
	return result, nil
}
