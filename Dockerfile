ARG PREFIX=golang
FROM ${PREFIX}:1.9.2-alpine3.6

ARG ARCH=x86_64
ENV arch=${ARCH}
ENV username dockerbuild
ENV userid 1000
ENV groupid 1000
ENV homedir /home/${username}
ENV pkgsdir ${homedir}/packages/main
ENV aportsdir /tmp/aports/main
ENV apkopts --allow-untrusted
ENV distfiles /var/cache/distfiles
ENV buildjobs 9

ADD aports.patch /tmp/aports.patch

RUN apk add --no-cache \
    alpine-sdk \
    autoconf \
    automake \
    sed \
    xz \
    intltool \
    shared-mime-info \
    hicolor-icon-theme \
    gtk-update-icon-cache \
    curl-dev \
    gettext-static \
    python2-dev \
    tiff-dev \
    libjpeg-turbo-dev \
    linux-headers \
    ncurses-dev \
    libcap-ng-dev \
    cups-dev \
    gnutls-dev \
    gtk-doc \
    libice-dev \
    libxcomposite-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxi-dev \
    gobject-introspection-dev \
    libxdamage-dev \
    icu-dev \
    libxft-dev \
    && echo ""

RUN printf "%s:x:%d:%d:Linux User,,,:%s:%s" ${username} ${userid} \
        ${groupid} ${homedir} "/bin/ash" >> /etc/passwd && \
    mkdir -p ${homedir} && \
    chown -R ${userid}:${groupid} ${homedir} && \
    addgroup ${username} abuild && \
    mkdir -p ${distfiles} && \
    chmod a+w ${distfiles} && \
    echo ${username} ALL=\(ALL\) ALL >> /etc/sudoers && \
    sed "s/export JOBS=.*/export JOBS=${buildjobs}/g" \
        -i /etc/abuild.conf && \
    chown -R ${userid}:${groupid} /usr/local/go

USER ${username}
RUN abuild-keygen -a

USER root
RUN cp -p ${homedir}/.abuild/*.rsa.pub /etc/apk/keys

USER ${username}
RUN cd /tmp && \
    git clone \
        git://git.alpinelinux.org/aports \
        --depth=1 \
        --branch=v3.6.2 && \
    cd aports && \
    patch -p1 < ../aports.patch

USER ${username}
RUN cd ${aportsdir}/atk && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/atk-*.apk

USER ${username}
RUN cd ${aportsdir}/cairo && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/cairo-*.apk

USER ${username}
RUN cd ${aportsdir}/fontconfig && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/fontconfig-*.apk

USER ${username}
RUN cd ${aportsdir}/freetype && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/freetype-*.apk

USER ${username}
RUN cd ${aportsdir}/harfbuzz && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/harfbuzz-*.apk

USER ${username}
RUN cd ${aportsdir}/pango && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/pango-*.apk

USER ${username}
RUN cd ${aportsdir}/pixman && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/pixman-*.apk
USER ${username}

RUN cd ${aportsdir}/util-linux && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/util-linux-*.apk
USER ${username}

RUN cd ${aportsdir}/gdk-pixbuf && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/gdk-pixbuf-*.apk
USER ${username}

RUN cd ${aportsdir}/glib && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/glib-*.apk
USER ${username}

RUN cd ${aportsdir}/gtk+2.0 && abuild
USER root
RUN apk add ${apkopts} ${pkgsdir}/*/gtk+2.0-*.apk

USER ${username}
CMD setarch ${arch} ./release.sh
