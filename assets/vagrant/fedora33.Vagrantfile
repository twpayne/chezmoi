Vagrant.configure("2") do |config|
  config.vm.box = "generic/fedora33"
  config.vm.define :fedora33
  config.vm.hostname = "fedora33"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    dnf install --quiet --assumeyes bzip2 git gnupg golang
  SHELL
  config.vm.provision "file", source: "assets/vagrant/fedora33.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
