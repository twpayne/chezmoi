Vagrant.configure("2") do |config|
  config.vm.box = "openindiana/hipster"
  config.vm.box_version = "20220105"
  config.vm.synced_folder ".", "/chezmoi", type: "rsync"
  config.vm.provision "shell", inline: <<-SHELL
    pkg install -q compress/zip developer/gcc-7 developer/golang developer/versioning/git
  SHELL
  config.vm.provision "shell", inline: <<-SHELL
    echo export CHEZMOI_GITHUB_ACCESS_TOKEN=#{ENV['CHEZMOI_GITHUB_ACCESS_TOKEN']} >> /export/home/vagrant/.profile
    echo export CHEZMOI_GITHUB_TOKEN=#{ENV['CHEZMOI_GITHUB_TOKEN']} >> /export/home/vagrant/.profile
    echo export GITHUB_ACCESS_TOKEN=#{ENV['GITHUB_ACCESS_TOKEN']} >> /export/home/vagrant/.profile
    echo export GITHUB_TOKEN=#{ENV['GITHUB_TOKEN']} >> /export/home/vagrant/.profile
  SHELL
  config.vm.provision "file", source: "assets/vagrant/openindiana.test-chezmoi.sh", destination: "test-chezmoi.sh"
end
