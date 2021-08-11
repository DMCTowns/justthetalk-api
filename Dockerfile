# This file is part of the JUSTtheTalkAPI distribution (https://github.com/jdudmesh/justthetalk-api).
# Copyright (c) 2021 John Dudmesh.

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, version 3.

# This program is distributed in the hope that it will be useful, but
# WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
# General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

FROM golang

ENV DB_HOST=localhost
ENV DB_PORT=3306
ENV REDIS_HOST=localhost
ENV REDIS_PORT=6379
ENV PLATFORM=PRODUCTION

RUN mkdir -p /go/src/justthetalk
WORKDIR /go/src/justthetalk
COPY . .

RUN go build

CMD ["./justthetalk"]
