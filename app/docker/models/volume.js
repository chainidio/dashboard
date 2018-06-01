function VolumeViewModel(data) {
  this.Id = data.Name;
  this.Driver = data.Driver;
  this.Options = data.Options;
  this.Labels = data.Labels;
  if (this.Labels && this.Labels['com.docker.compose.project']) {
    this.StackName = this.Labels['com.docker.compose.project'];
  } else if (this.Labels && this.Labels['com.docker.stack.namespace']) {
    this.StackName = this.Labels['com.docker.stack.namespace'];
  }
  this.Mountpoint = data.Mountpoint;

  if (data.Chain Platform) {
    if (data.Chain Platform.ResourceControl) {
      this.ResourceControl = new ResourceControlViewModel(data.Chain Platform.ResourceControl);
    }
    if (data.Chain Platform.Agent && data.Chain Platform.Agent.NodeName) {
      this.NodeName = data.Chain Platform.Agent.NodeName;
    }
  }
}
