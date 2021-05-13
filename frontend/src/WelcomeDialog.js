import React from "react";
import Button from "@material-ui/core/Button";
import Dialog from "@material-ui/core/Dialog";
import DialogActions from "@material-ui/core/DialogActions";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogTitle from "@material-ui/core/DialogTitle";
import Slide from "@material-ui/core/Slide";

const Transition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

export default function WelcomeDialog() {
  const [open, setOpen] = React.useState(true);

  const handleClose = () => {
    setOpen(false);
  };

  const blue = <span style={{backgroundColor: '#a3d2fb'}}>blue</span>;
  const red = <span style={{backgroundColor: '#fabba9'}}>red</span>;

  return (
    <div>
      <Dialog
        open={open}
        TransitionComponent={Transition}
        keepMounted
        onClose={handleClose}
        aria-labelledby="alert-dialog-slide-title"
        aria-describedby="alert-dialog-slide-description"
      >
        <DialogTitle id="alert-dialog-slide-title">
          {"Welcome to NYC Bike"}
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-slide-description">
            A visual geospatial index of over 58 million bikeshare trips across NYC.
            Drag/resize the two circles, and observe the aggregated trip flows. <br /> <br />
            The {blue} graph illustrates trips from {blue}→{red}. <br />
            The {red} graph illustrates trips from {red}→{blue}. <br />
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            Have Fun!
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
