class Lockbox < Formula
  desc "Team GPG key management for git repositories"
  homepage "https://github.com/yourusername/lockbox"
  url "https://github.com/yourusername/lockbox/archive/v0.1.0.tar.gz"
  sha256 "..."

  depends_on "python@3.12"
  depends_on "gpgme"
  depends_on "cython"

  def install
    venv = virtualenv_create(libexec, "python3.12")
    venv.pip_install "cython"
    venv.pip_install_and_link buildpath
  end

  test do
    system bin/"lockbox", "--version"
  end
end 