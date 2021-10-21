Vagrant.configure("2") do |config|
  config.vm.box = "openindiana/hipster"
  config.vm.box_version = "202109"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    pkg install -q compress/zip developer/gcc-7 developer/golang developer/versioning/git
  SHELL
  config.vm.provision "file", source: "assets/vagrant/openindiana.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
