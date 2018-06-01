function EndpointGroupDefaultModel() {
  this.Name = '';
  this.Description = '';
  this.Labels = [];
}

function EndpointGroupModel(data) {
  this.Id = data.Id;
  this.Name = data.Name;
  this.Description = data.Description;
  this.Labels = data.Labels;
  this.AuthorizedUsers = data.AuthorizedUsers;
  this.AuthorizedTeams = data.AuthorizedTeams;
}

function EndpointGroupCreateRequest(model, endpoints) {
  this.Name = model.Name;
  this.Description = model.Description;
  this.Labels = model.Labels;
  this.AssociatedEndpoints = endpoints;
}

function EndpointGroupUpdateRequest(model, endpoints) {
  this.id = model.Id;
  this.Name = model.Name;
  this.Description = model.Description;
  this.Labels = model.Labels;
  this.AssociatedEndpoints = endpoints;
}
