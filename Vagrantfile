Vagrant.configure(2) do |config|
  config.vm.box = "centos/7"
  config.vm.provision "shell", path: "vagrant.sh"
  config.vm.provider "virtualbox" do |vb|
    vb.memory = "1024"
  end
end
