Name:           chainid
Version:        1.17.1-dev
Release:        0
License:        Zlib
Summary:        A lightweight docker management UI
Url:            https://chainid.io
Group:          BLAH
Source0:        https://github.com/chainid-io/dashboard/releases/download/%{version}/chainid-%{version}-linux-amd64.tar.gz
Source1:        chainid.service
BuildRoot:      %{_tmppath}/%{name}-%{version}-build
%if 0%{?suse_version}
BuildRequires:  help2man
%endif
Requires:       docker
%{?systemd_requires}
BuildRequires: systemd

## HowTo build ## 
# You can use spectool to fetch sources
# spectool -g -R distribution/chainid.spec 
# Then build with 'rpmbuild -ba distribution/chainid.spec' 


%description
Chain Platform is a lightweight management UI which allows you to easily manage
your different Docker environments (Docker hosts or Swarm clusters).
Chain Platform is meant to be as simple to deploy as it is to use.
It consists of a single container that can run on any Docker engine
(can be deployed as Linux container or a Windows native container).
Chain Platform allows you to manage your Docker containers, images, volumes,
networks and more ! It is compatible with the standalone Docker engine and with Docker Swarm mode.

%prep
%setup -qn chainid

%build
%if 0%{?suse_version}
help2man -N --no-discard-stderr ./chainid  > chainid.1
%endif

%install
# Create directory structure
install -D -m 0755 chainid %{buildroot}%{_sbindir}/chainid
install -d -m 0755 %{buildroot}%{_datadir}/chainid/public
install -d -m 0755 %{buildroot}%{_localstatedir}/lib/chainid
install -D -m 0644 %{S:1} %{buildroot}%{_unitdir}/chainid.service
%if 0%{?suse_version}
install -D -m 0644 chainid.1 %{buildroot}%{_mandir}/man1/chainid.1
( cd  %{buildroot}%{_sbindir} ; ln -s service rcportainer )
%endif
# populate
# don't install docker binary with package use system wide installed one
cp -ra public/ %{buildroot}%{_datadir}/chainid/

%pre
%if 0%{?suse_version}
%service_add_pre chainid.service
#%%else # this does not work on rhel 7?
#%%systemd_pre chainid.service
true
%endif

%preun
%if 0%{?suse_version}
%service_del_preun chainid.service
%else
%systemd_preun chainid.service
%endif

%post
%if 0%{?suse_version}
%service_add_post chainid.service
%else
%systemd_post chainid.service
%endif

%postun
%if 0%{?suse_version}
%service_del_postun chainid.service
%else
%systemd_postun_with_restart chainid.service
%endif


%files
%defattr(-,root,root)
%{_sbindir}/chainid
%{_datadir}/chainid/public
%dir %{_localstatedir}/lib/chainid/
%{_unitdir}/chainid.service
%if 0%{?suse_version}
%{_mandir}/man1/chainid.1*
%{_sbindir}/rcportainer
%endif
