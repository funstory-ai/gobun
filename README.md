<div align="center">
    <img src="https://github.com/funstory-ai/gobun/blob/main/uglylogo.png?raw=true" alt="GoBun Logo" width="200">
  <p>Develop, Train, and Scale AI Applications Serverlessly and Cheaply</p>
</div>

## What is GoBun?

GoBun is built for connecting consumer grade gpu cloud providers and unifying them into a simple, uniformed and production-ready service.

## Why use GoBun?

- **Cheap**: start from 0.3 USD/GPU/hour and pay as you go in minutes
- **Easy**: simple to use, easy to deploy, easy to scale
- **Production Ready**: GoBun provide a daemon to recover from the failure.
- **Serverless**: no need to manage servers, no need to worry about the underlying infrastructure, 

## Getting Started

1. Register an account in XianGongYun
2. Get the token from [XianGongYun](https://www.xiangongyun.com/console/user/accesstoken)
3. Set the token to environment variable `XGY_TOKEN`
4. Run `gobun create` to create a pod
5. Run `gobun list --watch` to watch the pod status
6. When the pod is running, run `gobun attach <pod_id>` to attach to the pod
