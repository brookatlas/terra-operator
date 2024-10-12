package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"log"
	"net/http"
	"os"
	"reflect"
)

type TerraformRequest struct {
	Version    string            `json:"Version"`
	ModulePath string            `json:"ModulePath"`
	Variables  map[string]string `json:"Variables"`
}

var installedVersions = make(map[string]string)

func installTerraform(versionString string) error {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(versionString)),
	}
	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing terraform: %s", err)
		return err
	}
	installedVersions[versionString] = execPath
	return nil
}

func terraformPlan(c *gin.Context) {
	var request TerraformRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema"})
		return
	}
	// install terraform
	execPath, installed := installedVersions[request.Version]
	if !installed {
		err := installTerraform(request.Version)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Fatalf("error installing terraform: %s", err)
			return
		}
	}
	execPath, _ = installedVersions[request.Version]
	// clone the repository
	dir, err := os.MkdirTemp("", "cloned-terraform-module")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error creating a new temporary directory"})
		log.Fatalf("error creating a new temporary directory: %s", err)
		return
	}
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:      request.ModulePath,
		Progress: os.Stdout,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error cloning the repository"})
		log.Fatalf("error cloning repository: %s", err)
		return
	}
	err = os.Chdir(dir)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error moving to repo directory"})
		log.Fatalf("error moving to repo directory: %s", err)
		return
	}
	// run terraform plan with variables and output to json
	terraform, err := tfexec.NewTerraform(dir, execPath)
	var args []reflect.Value
	var planBuff bytes.Buffer
	writer := bufio.NewWriter(&planBuff)
	args = append(args, reflect.ValueOf(context.Background()))
	args = append(args, reflect.ValueOf(writer))
	for key, value := range request.Variables {
		assignmentString := fmt.Sprintf("%s:%s", key, value)
		args = append(args, reflect.ValueOf(tfexec.Var(assignmentString)))
	}
	err = terraform.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error running init"})
		log.Fatalf("error running init: %s", err)
		return
	}
	planJsonFunc := reflect.ValueOf(terraform.PlanJSON)
	results := planJsonFunc.Call(args)
	_ = results[0].Interface().(bool)
	if results[1].IsNil() {
		err = nil
	} else {
		err = results[1].Interface().(error)
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error running init"})
		log.Fatalf("error running init: %s", err)
		return
	}
	fmt.Println("plan json output: \n", writer.re)
}

//func terraformApply(c *gin.Context) {
//	var request TerraformRequest
//	if err := c.ShouldBindJSON(&request); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schema"})
//		return
//	}
//
//	// clone the repository
//	// run terraform apply
//	// save apply output metadata in configmap
//}

func main() {
	app := gin.Default()
	app.POST("/plan", terraformPlan)
	//app.POST("/apply", terraformApply)
	err := app.Run(":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
