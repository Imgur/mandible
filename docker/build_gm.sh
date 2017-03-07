#!/bin/bash

apt-get install -y libjpeg-dev liblcms2-dev libwmf-dev libx11-dev libsm-dev libice-dev libxext-dev x11proto-core-dev libxml2-dev libfreetype6-dev libexif-dev libbz2-dev libtiff-dev libjbig-dev zlib1g-dev libpng-dev libwebp-dev ghostscript gsfonts autotools-dev transfig sharutils libltdl-dev mercurial cmake
wget "http://www.ece.uvic.ca/~frodo/jasper/software/jasper-2.0.12.tar.gz" -O jasper.tar.gz
mkdir jasper && tar -xvzf jasper.tar.gz -C jasper --strip-components 1 && cd jasper
mkdir BUILD && cd BUILD && cmake -DCMAKE_INSTALL_PREFIX=/usr    \
      -DCMAKE_BUILD_TYPE=Release     \
      -DCMAKE_SKIP_INSTALL_RPATH=YES \
      -DCMAKE_INSTALL_DOCDIR=/usr/share/doc/jasper-2.0.10 \
      ..  &&
make
make install
cd ../.. && rm -rf jasper jasper.tar.gz
hg clone http://hg.code.sf.net/p/graphicsmagick/code GM
cd GM
hg update -r tip
CC="gcc" CFLAGS="-fopenmp -Wall -g -fno-strict-aliasing -O2 -Wall -pthread" CPPFLAGS="-I/usr/include/X11 -I/usr/include/freetype2 -I/usr/include/libxml2" CXX="g++" CXXFLAGS="-Wall -g -fno-strict-aliasing -O2 -pthread" LDFLAGS="-L/usr/lib/X11 -L/usr/lib/x86_64-linux-gnu" LIBS="-ljbig -lwebp -llcms2 -ltiff -lfreetype -ljpeg -lpng16 -lwmflite -lXext -lSM -lICE -lX11 -llzma -lbz2 -lxml2 -lz -lm -lgomp -lpthread" ./configure  '--build' 'x86_64-linux-gnu' '--enable-shared' '--enable-static' '--enable-libtool-verbose' '--prefix=/usr' '--mandir=${prefix}/share/man' '--infodir=${prefix}/share/info' '--docdir=${prefix}/share/doc/graphicsmagick' '--with-gs-font-dir=/usr/share/fonts/type1/gsfonts' '--with-x' '--x-includes=/usr/include/X11' '--x-libraries=/usr/lib/X11' '--without-dps' '--without-modules' '--without-frozenpaths' '--with-webp' '--with-perl' '--with-perl-options=INSTALLDIRS=vendor' '--enable-quantum-library-names' '--with-quantum-depth=16' 'build_alias=x86_64-linux-gnu' 'CFLAGS=-Wall -g -fno-strict-aliasing -O2' 'LDFLAGS=' 'CXXFLAGS=-Wall -g -fno-strict-aliasing -O2'
make
make install
cd .. && rm -rf GM
