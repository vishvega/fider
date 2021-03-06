import "./SupportCounter.scss";

import * as React from "react";
import { Post, PostStatus } from "@fider/models";
import { actions, device, classSet, Fider } from "@fider/services";
import { SignInModal } from "@fider/components";

interface SupportCounterProps {
  post: Post;
}

interface SupportCounterState {
  supported: boolean;
  total: number;
  showSignIn: boolean;
}

export class SupportCounter extends React.Component<SupportCounterProps, SupportCounterState> {
  constructor(props: SupportCounterProps) {
    super(props);
    this.state = {
      supported: props.post.viewerSupported,
      total: props.post.totalSupporters,
      showSignIn: false
    };
  }

  private supportOrUndo = async () => {
    if (!Fider.session.isAuthenticated) {
      this.setState({ showSignIn: true });
      return;
    }

    const action = this.state.supported ? actions.removeSupport : actions.addSupport;

    const response = await action(this.props.post.number);
    if (response.ok) {
      this.setState(state => ({
        supported: !state.supported,
        total: state.total + (state.supported ? -1 : 1)
      }));
    }
  };

  public render() {
    const status = PostStatus.Get(this.props.post.status);

    const className = classSet({
      "m-supported": !status.closed && this.state.supported,
      "m-disabled": status.closed,
      "no-touch": !device.isTouch()
    });

    const vote = (
      <button className={className} onClick={this.supportOrUndo}>
        <i className="caret up icon" />
        {this.state.total}
      </button>
    );

    const disabled = (
      <button className={className}>
        <i className="caret up icon" />
        {this.state.total}
      </button>
    );

    return (
      <>
        <SignInModal isOpen={this.state.showSignIn} />
        <div className="c-support-counter">{status.closed ? disabled : vote}</div>
      </>
    );
  }
}
