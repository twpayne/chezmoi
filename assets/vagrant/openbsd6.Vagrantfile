Vagrant.configure("2") do |config|
  config.vm.box = "generic/openbsd6"
  config.vm.hostname = "openbsd6"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    pkg_add -x git gnupg go
  SHELL
  config.vm.provision "file", source: "assets/vagrant/openbsd6.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
