package conf

import (
    "fmt"
    "github.com/go-ini/ini"
)

type AnsibleInventory struct {
    ini.File
}

func NewAnsibleInventory() (AnsibleInventory, error) {
    config := NewConfig()
    cfg, err := ini.LoadSources(ini.LoadOptions{
        // 允许布尔键存在，ansible_hosts文件往往IP都是独立一个没有键，兼容这种情况
        AllowBooleanKeys: true,
    }, config.Ansible.HostsFile)
    if err != nil {
        return AnsibleInventory{}, err
    }
    return AnsibleInventory{
        File: *cfg,
    }, nil
}

type HostItem struct {
    Title    string     `json:"title"`
    Key      string     `json:"key"`
    Children []HostItem `json:"children"`
}

// Like this:
// {type: "local", key: "local", children: [{type: "local", key: "local"}]}
func (ai *AnsibleInventory) GetHostsAll() ([]HostItem, error) {
    var hostItems []HostItem
    sections := ai.GetSections()

    for sIndex, section := range sections {
        hosts, err := ai.GetHosts(section)
        if err != nil {
            return nil, err
        }
        var sectionItems []HostItem
        for hIndex, host := range hosts {
            sectionItems = append(sectionItems, HostItem{
                Key:   fmt.Sprintf("%d-%d", sIndex, hIndex),
                Title: host,
            })
        }
        hostItems = append(hostItems, HostItem{
            Key:      section,
            Title:    section,
            Children: sectionItems,
        })
    }
    return hostItems, nil
}

func (ai *AnsibleInventory) GetSections() []string {
    sections := ai.SectionStrings()
    // remove "DEFAULT"
    for i, s := range sections {
        if s == "DEFAULT" {
            sections = append(sections[:i], sections[i+1:]...)
        }
    }
    return sections
}

func (ai *AnsibleInventory) GetHosts(sectionName string) ([]string, error) {
    section, err := ai.GetSection(sectionName)
    if err != nil {
        return nil, err
    }

    hostKeys := section.Keys()
    var hosts []string
    for _, k := range hostKeys {
        if k.Value() == "true" {
            hosts = append(hosts, k.Name())
        }
    }

    return hosts, nil
}
