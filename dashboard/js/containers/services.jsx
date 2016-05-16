import React, { PropTypes, Component } from 'react'
import { connect } from 'react-redux'
import Service from '../components/service'
import { setServiceTimeout, toggleServiceChecked } from '../actions/'

function compare(a,b) {
  if (a.name < b.name)
    return -1;
  else if (a.name > b.name)
    return 1;
  else
    return 0;
}

class Services extends Component {

  render() {
    console.log("RENDERRENDER", this.props)
    const services = this.props.services
    services.sort(compare)
    return (<ul className="services">
      {services.map(service => <Service key={service.name} service={service} checked={this.props.checked[service.name] || false}
        toggleChecked={() => this.props.toggleServiceChecked(this.props.viewId, service.name)}
        setServiceTimeout={this.props.setServiceTimeout}/>)}
    </ul>)
  }
}

Services.propTypes = {
  viewId: PropTypes.string.isRequired,
  services: PropTypes.array.isRequired,
  checked: PropTypes.object.isRequired,
  setServiceTimeout: PropTypes.func.isRequired,
  toggleServiceChecked: PropTypes.func.isRequired
}

function mapStateToProps(state) {
  return {}
}

export default connect(mapStateToProps, { setServiceTimeout, toggleServiceChecked })(Services)