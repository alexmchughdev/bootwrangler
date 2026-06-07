package windows

// autounattendTmpl is a UEFI/GPT Windows Setup answer file. Dynamic text is
// XML-escaped via the {{ xml }} template function.
const autounattendTmpl = `<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
  <settings pass="windowsPE">
    <component name="Microsoft-Windows-International-Core-WinPE" processorArchitecture="{{ .Arch }}" language="neutral" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <SetupUILanguage><UILanguage>{{ xml .Locale }}</UILanguage></SetupUILanguage>
      <InputLocale>{{ xml .Locale }}</InputLocale>
      <SystemLocale>{{ xml .Locale }}</SystemLocale>
      <UILanguage>{{ xml .Locale }}</UILanguage>
      <UserLocale>{{ xml .Locale }}</UserLocale>
    </component>
    <component name="Microsoft-Windows-Setup" processorArchitecture="{{ .Arch }}" language="neutral" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <DiskConfiguration>
        <Disk wcm:action="add">
          <DiskID>0</DiskID>
          <WillWipeDisk>true</WillWipeDisk>
          <CreatePartitions>
            <CreatePartition wcm:action="add"><Order>1</Order><Type>EFI</Type><Size>260</Size></CreatePartition>
            <CreatePartition wcm:action="add"><Order>2</Order><Type>MSR</Type><Size>16</Size></CreatePartition>
            <CreatePartition wcm:action="add"><Order>3</Order><Type>Primary</Type><Extend>true</Extend></CreatePartition>
          </CreatePartitions>
          <ModifyPartitions>
            <ModifyPartition wcm:action="add"><Order>1</Order><PartitionID>1</PartitionID><Format>FAT32</Format><Label>System</Label></ModifyPartition>
            <ModifyPartition wcm:action="add"><Order>2</Order><PartitionID>2</PartitionID></ModifyPartition>
            <ModifyPartition wcm:action="add"><Order>3</Order><PartitionID>3</PartitionID><Format>NTFS</Format><Label>Windows</Label></ModifyPartition>
          </ModifyPartitions>
        </Disk>
      </DiskConfiguration>
      <ImageInstall>
        <OSImage>
          <InstallFrom>
            <MetaData wcm:action="add"><Key>/IMAGE/NAME</Key><Value>{{ xml .Edition }}</Value></MetaData>
          </InstallFrom>
          <InstallTo><DiskID>0</DiskID><PartitionID>3</PartitionID></InstallTo>
        </OSImage>
      </ImageInstall>
      <UserData>
        <AcceptEula>true</AcceptEula>
        {{ if .ProductKey }}<ProductKey><Key>{{ xml .ProductKey }}</Key></ProductKey>{{ end }}
        {{ if .Owner }}<FullName>{{ xml .Owner }}</FullName>{{ end }}
        {{ if .Organization }}<Organization>{{ xml .Organization }}</Organization>{{ end }}
      </UserData>
    </component>
  </settings>
{{ if eq .JoinMethod "offline-djoin" }}
  <settings pass="offlineServicing">
    <component name="Microsoft-Windows-UnattendedJoin" processorArchitecture="{{ .Arch }}" language="neutral" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <OfflineIdentification>
        <Provisioning>
          <AccountData>{{ xml .ProvisionBlob }}</AccountData>
        </Provisioning>
      </OfflineIdentification>
    </component>
  </settings>
{{ end }}
  <settings pass="specialize">
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="{{ .Arch }}" language="neutral" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <ComputerName>{{ xml .ComputerName }}</ComputerName>
      {{ if .Organization }}<RegisteredOrganization>{{ xml .Organization }}</RegisteredOrganization>{{ end }}
      {{ if .Owner }}<RegisteredOwner>{{ xml .Owner }}</RegisteredOwner>{{ end }}
    </component>
{{ if eq .JoinMethod "unattend" }}
    <component name="Microsoft-Windows-UnattendedJoin" processorArchitecture="{{ .Arch }}" language="neutral" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <Identification>
        <JoinDomain>{{ xml .Domain }}</JoinDomain>
        {{ if .OU }}<MachineObjectOU>{{ xml .OU }}</MachineObjectOU>{{ end }}
        <Credentials>
          <Domain>{{ xml .Domain }}</Domain>
          <Username>{{ xml .JoinUser }}</Username>
          <!-- Fill in the join password before deploying; BootWrangler never stores it. -->
          <Password></Password>
        </Credentials>
      </Identification>
    </component>
{{ end }}
  </settings>
  <settings pass="oobeSystem">
    <component name="Microsoft-Windows-International-Core" processorArchitecture="{{ .Arch }}" language="neutral" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <InputLocale>{{ xml .Locale }}</InputLocale>
      <SystemLocale>{{ xml .Locale }}</SystemLocale>
      <UILanguage>{{ xml .Locale }}</UILanguage>
      <UserLocale>{{ xml .Locale }}</UserLocale>
    </component>
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="{{ .Arch }}" language="neutral" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State">
      <OOBE>
        <HideEULAPage>true</HideEULAPage>
        <HideOEMRegistrationScreen>true</HideOEMRegistrationScreen>
        <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
        <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
        <ProtectYourPC>3</ProtectYourPC>
      </OOBE>
      <UserAccounts>
        <LocalAccounts>
          <LocalAccount wcm:action="add">
            <Name>{{ xml .AdminUser }}</Name>
            <Group>Administrators</Group>
            <DisplayName>{{ xml .AdminUser }}</DisplayName>
            <Password><Value>{{ xml .AdminPassword }}</Value><PlainText>true</PlainText></Password>
          </LocalAccount>
        </LocalAccounts>
      </UserAccounts>
      <AutoLogon>
        <Enabled>true</Enabled>
        <Username>{{ xml .AdminUser }}</Username>
        <LogonCount>1</LogonCount>
        <Password><Value>{{ xml .AdminPassword }}</Value><PlainText>true</PlainText></Password>
      </AutoLogon>
      <FirstLogonCommands>
        <SynchronousCommand wcm:action="add">
          <Order>1</Order>
          <CommandLine>powershell.exe -ExecutionPolicy Bypass -NoProfile -File C:\Windows\Setup\Scripts\FirstLogon.ps1</CommandLine>
          <Description>BootWrangler first-logon provisioning</Description>
        </SynchronousCommand>
      </FirstLogonCommands>
    </component>
  </settings>
</unattend>
`

// firstLogonTmpl runs once at first logon. Delivered via the install media's
// sources\$OEM$\$$\Setup\Scripts\ folder, which Setup copies to
// C:\Windows\Setup\Scripts\.
const firstLogonTmpl = `# BootWrangler first-logon provisioning
$ErrorActionPreference = "Continue"
Start-Transcript -Path "$env:SystemRoot\Temp\bootwrangler-firstlogon.log" -Append

{{ if .EnableSSH }}
# --- OpenSSH server ---
Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
Set-Service -Name sshd -StartupType Automatic
Start-Service sshd
New-NetFirewallRule -Name sshd -DisplayName "OpenSSH Server (sshd)" -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22 -ErrorAction SilentlyContinue
{{ end }}
{{ if .Packages }}
# --- Packages (winget) ---
{{ range .Packages }}winget install --silent --accept-package-agreements --accept-source-agreements --id {{ . }}
{{ end }}{{ end }}
{{ if eq .JoinMethod "first-logon" }}
# --- Active Directory domain join ---
# Supply credentials for {{ .JoinUser }} at deploy time. Replace Get-Credential
# with a secure source for fully unattended runs; nothing is stored by BootWrangler.
$cred = Get-Credential -UserName "{{ .Domain }}\{{ .JoinUser }}" -Message "Domain join: {{ .Domain }}"
Add-Computer -DomainName "{{ .Domain }}"{{ if .OU }} -OUPath "{{ .OU }}"{{ end }} -Credential $cred -Force -Restart
{{ end }}
{{ range .Scripts }}
# --- {{ .Name }} ---
{{ .Content }}
{{ end }}
Stop-Transcript
`
