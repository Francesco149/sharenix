/*
   Copyright 2014 Franc[e]sco (lolisamurai@tfwno.gf)
   This file is part of sharenix.
   sharenix is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   sharenix is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with sharenix. If not, see <http://www.gnu.org/licenses/>.
*/

package sharenixlib

import "fmt"

// A NotImplementedError is returned when the called feature is not implemented
type NotImplementedError struct{}

func (e *NotImplementedError) Error() string {
	return "Not implemented!"
}

// A SiteNotFoundError is returned when the target site doesn't exist in the config
type SiteNotFoundError struct {
	site string
}

func (e *SiteNotFoundError) Error() string {
	return fmt.Sprintf("Site not found: %s", e.site)
}
